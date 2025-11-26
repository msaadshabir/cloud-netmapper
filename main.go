package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cloud-netmapper/config"
	"cloud-netmapper/logger"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Command line flags
var (
	regionFlag     string
	allRegionsFlag bool
	outputFormat   string
	outputDir      string
	verbosity      string
	configFile     string
	diffMode       bool
	saveSnapshot   bool
)

func init() {
	flag.StringVar(&regionFlag, "region", "", "AWS region to scan (comma-separated for multiple)")
	flag.BoolVar(&allRegionsFlag, "all-regions", false, "Scan all available AWS regions")
	flag.StringVar(&outputFormat, "format", "", "Output format: png, svg, json, csv, markdown, html, sarif")
	flag.StringVar(&outputDir, "output-dir", "", "Output directory for generated files")
	flag.StringVar(&verbosity, "verbosity", "", "Log verbosity: debug, info, warn, error")
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.BoolVar(&diffMode, "diff", false, "Compare with previous scan and show changes")
	flag.BoolVar(&saveSnapshot, "save-snapshot", true, "Save snapshot for future diff comparisons")
}

// getAvailableRegions fetches all available AWS regions
func getAvailableRegions() ([]string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	resp, err := ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %v", err)
	}

	var regions []string
	for _, region := range resp.Regions {
		if region.RegionName != nil {
			regions = append(regions, *region.RegionName)
		}
	}
	return regions, nil
}

func main() {
	flag.Parse()

	// Load configuration file
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override config with command line flags (flags take precedence)
	if verbosity == "" {
		verbosity = "info"
	}
	logger.Init(logger.ParseLevel(verbosity))

	// Use config file values as defaults, override with CLI flags
	if outputFormat == "" {
		outputFormat = cfg.Output.Format
	}
	if outputDir == "" {
		outputDir = cfg.Output.Directory
	}
	// Ensure outputDir has a valid default
	if outputDir == "" {
		outputDir = "."
	}

	// Determine which regions to scan
	var regions []string
	if allRegionsFlag {
		logger.Info("Fetching available AWS regions...")
		regions, err = getAvailableRegions()
		if err != nil {
			logger.Error("Failed to get regions", "error", err)
			os.Exit(1)
		}
		logger.Info("Found regions", "count", len(regions))
	} else if regionFlag != "" {
		regions = strings.Split(regionFlag, ",")
		for i := range regions {
			regions[i] = strings.TrimSpace(regions[i])
		}
	} else if len(cfg.Regions) > 0 {
		regions = cfg.Regions
	} else {
		regions = []string{"us-east-1"}
	}

	// Create output directory if it doesn't exist
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			logger.Error("Failed to create output directory", "error", err)
			os.Exit(1)
		}
	}

	// Collect resources from all regions
	allResources := make(map[string]*AWSResources)
	var allRisks []Risk

	for _, region := range regions {
		logger.Info("Scanning region", "region", region)
		resources, err := getAWSResources(region)
		if err != nil {
			logger.Warn("Failed to scan region", "region", region, "error", err)
			continue
		}
		allResources[region] = resources

		// Check security risks for this region
		risks := checkSecurityRisks(resources)
		for i := range risks {
			risks[i].Resource = fmt.Sprintf("[%s] %s", region, risks[i].Resource)
		}
		allRisks = append(allRisks, risks...)
	}

	if len(allResources) == 0 {
		logger.Error("No resources collected from any region")
		os.Exit(1)
	}

	// Merge resources for output
	mergedResources := mergeResources(allResources)

	// Handle diff mode - compare with previous scan
	if diffMode {
		for region, resources := range allResources {
			logger.Info("Loading previous snapshot for diff", "region", region)
			prevSnapshot, err := LoadPreviousSnapshot(region, outputDir)
			if err != nil {
				logger.Warn("Failed to load previous snapshot", "region", region, "error", err)
				continue
			}

			if prevSnapshot == nil {
				logger.Info("No previous snapshot found for region", "region", region)
				continue
			}

			diffReport := DetectChanges(resources, prevSnapshot.Resources)
			diffReport.PreviousScan = prevSnapshot.Timestamp

			if diffReport.Summary.Total > 0 {
				fmt.Printf("\n=== Changes detected in %s ===\n", region)
				fmt.Printf("Added: %d, Removed: %d, Modified: %d\n",
					diffReport.Summary.Added, diffReport.Summary.Removed, diffReport.Summary.Modified)

				for _, change := range diffReport.Changes {
					symbol := "?"
					switch change.ChangeType {
					case "added":
						symbol = "+"
					case "removed":
						symbol = "-"
					case "modified":
						symbol = "~"
					}
					fmt.Printf("  [%s] %s: %s (%s) - %s\n",
						symbol, change.ResourceType, change.ResourceName, change.ResourceID, change.Details)
				}

				// Save diff report
				diffFile := fmt.Sprintf("%s/diff_report_%s.md", outputDir, region)
				if err := GenerateDiffReport(diffReport, diffFile); err != nil {
					logger.Warn("Failed to save diff report", "error", err)
				} else {
					logger.Info("Diff report saved", "path", diffFile)
				}
			} else {
				fmt.Printf("\nNo changes detected in %s since last scan.\n", region)
			}
		}
	}

	// Save snapshots for future diffs
	if saveSnapshot {
		for region, resources := range allResources {
			if err := SaveSnapshot(resources, region, outputDir); err != nil {
				logger.Warn("Failed to save snapshot", "region", region, "error", err)
			} else {
				logger.Debug("Snapshot saved", "region", region)
			}
		}
	}

	// Save raw data
	rawData, err := json.MarshalIndent(mergedResources, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal JSON", "error", err)
		os.Exit(1)
	}
	jsonPath := fmt.Sprintf("%s/aws_resources.json", outputDir)
	if err := os.WriteFile(jsonPath, rawData, 0644); err != nil {
		logger.Error("Failed to write file", "error", err)
		os.Exit(1)
	}
	logger.Info("Raw data saved", "path", jsonPath)

	// Generate DOT file
	dotFile := fmt.Sprintf("%s/network_map.dot", outputDir)
	if err := generateDOTFile(mergedResources, dotFile); err != nil {
		logger.Error("Failed to generate DOT file", "error", err)
		os.Exit(1)
	}

	// Generate output based on format
	switch outputFormat {
	case "png":
		pngFile := fmt.Sprintf("%s/network_map.png", outputDir)
		logger.Info("Rendering PNG diagram...")
		cmd := exec.Command("dot", "-Tpng", dotFile, "-o", pngFile)
		if err := cmd.Run(); err != nil {
			logger.Error("Failed to render PNG", "error", err)
			os.Exit(1)
		}
		logger.Info("Diagram saved", "path", pngFile)

	case "svg":
		svgFile := fmt.Sprintf("%s/network_map.svg", outputDir)
		logger.Info("Rendering SVG diagram...")
		cmd := exec.Command("dot", "-Tsvg", dotFile, "-o", svgFile)
		if err := cmd.Run(); err != nil {
			logger.Error("Failed to render SVG", "error", err)
			os.Exit(1)
		}
		logger.Info("Diagram saved", "path", svgFile)

	case "html":
		htmlFile := fmt.Sprintf("%s/network_map.html", outputDir)
		logger.Info("Generating interactive HTML visualization...")
		if err := generateHTMLFile(mergedResources, htmlFile); err != nil {
			logger.Error("Failed to generate HTML", "error", err)
			os.Exit(1)
		}
		logger.Info("Interactive visualization saved", "path", htmlFile)

	case "json":
		logger.Info("JSON output already saved", "path", jsonPath)

	case "markdown":
		mdFile := fmt.Sprintf("%s/report.md", outputDir)
		logger.Info("Generating Markdown report...")
		if err := generateMarkdownReport(mergedResources, allRisks, mdFile); err != nil {
			logger.Error("Failed to generate Markdown report", "error", err)
			os.Exit(1)
		}
		logger.Info("Markdown report saved", "path", mdFile)

	case "csv":
		csvFile := fmt.Sprintf("%s/resources.csv", outputDir)
		logger.Info("Generating CSV export...")
		if err := generateCSVReport(mergedResources, csvFile); err != nil {
			logger.Error("Failed to generate CSV", "error", err)
			os.Exit(1)
		}
		logger.Info("CSV export saved", "path", csvFile)

	case "sarif":
		sarifFile := fmt.Sprintf("%s/security_report.sarif", outputDir)
		logger.Info("Generating SARIF report...")
		if err := generateSARIFReport(allRisks, sarifFile); err != nil {
			logger.Error("Failed to generate SARIF report", "error", err)
			os.Exit(1)
		}
		logger.Info("SARIF report saved", "path", sarifFile)

	default:
		if outputFormat != "" {
			logger.Warn("Unknown output format, DOT file generated", "format", outputFormat)
		}
		logger.Info("DOT file generated", "path", dotFile)
	}

	// Print security risks
	if len(allRisks) > 0 {
		fmt.Println("\nSECURITY RISKS FOUND:")
		for _, risk := range allRisks {
			fmt.Printf("  - [%s] %s: %s (Resource: %s)\n",
				risk.Severity, risk.Type, risk.Details, risk.Resource)
		}
	} else {
		fmt.Println("\nNo critical security risks detected.")
	}

	// Print summary
	fmt.Printf("\nScan complete. Scanned %d region(s): %s\n", len(regions), strings.Join(regions, ", "))
}

// mergeResources combines resources from multiple regions into one
func mergeResources(resourcesByRegion map[string]*AWSResources) *AWSResources {
	merged := &AWSResources{}

	for region, resources := range resourcesByRegion {
		// Add region prefix to names for clarity
		for _, vpc := range resources.VPCs {
			vpc.Name = fmt.Sprintf("[%s] %s", region, vpc.Name)
			merged.VPCs = append(merged.VPCs, vpc)
		}
		for _, subnet := range resources.Subnets {
			subnet.Name = fmt.Sprintf("[%s] %s", region, subnet.Name)
			merged.Subnets = append(merged.Subnets, subnet)
		}
		for _, inst := range resources.Instances {
			inst.Name = fmt.Sprintf("[%s] %s", region, inst.Name)
			merged.Instances = append(merged.Instances, inst)
		}
		for _, sg := range resources.SecurityGroups {
			sg.Name = fmt.Sprintf("[%s] %s", region, sg.Name)
			merged.SecurityGroups = append(merged.SecurityGroups, sg)
		}
		for _, lb := range resources.LoadBalancers {
			lb.Name = fmt.Sprintf("[%s] %s", region, lb.Name)
			merged.LoadBalancers = append(merged.LoadBalancers, lb)
		}
		for _, rds := range resources.RDSInstances {
			rds.Name = fmt.Sprintf("[%s] %s", region, rds.Name)
			merged.RDSInstances = append(merged.RDSInstances, rds)
		}
		for _, nat := range resources.NATGateways {
			nat.Name = fmt.Sprintf("[%s] %s", region, nat.Name)
			merged.NATGateways = append(merged.NATGateways, nat)
		}
		for _, igw := range resources.InternetGateways {
			igw.Name = fmt.Sprintf("[%s] %s", region, igw.Name)
			merged.InternetGateways = append(merged.InternetGateways, igw)
		}
		for _, rt := range resources.RouteTables {
			rt.Name = fmt.Sprintf("[%s] %s", region, rt.Name)
			merged.RouteTables = append(merged.RouteTables, rt)
		}
		for _, peer := range resources.VPCPeerings {
			peer.Name = fmt.Sprintf("[%s] %s", region, peer.Name)
			merged.VPCPeerings = append(merged.VPCPeerings, peer)
		}
		for _, fn := range resources.LambdaFunctions {
			fn.Name = fmt.Sprintf("[%s] %s", region, fn.Name)
			merged.LambdaFunctions = append(merged.LambdaFunctions, fn)
		}
		for _, cluster := range resources.EKSClusters {
			cluster.Name = fmt.Sprintf("[%s] %s", region, cluster.Name)
			merged.EKSClusters = append(merged.EKSClusters, cluster)
		}
		for _, cluster := range resources.ECSClusters {
			cluster.Name = fmt.Sprintf("[%s] %s", region, cluster.Name)
			merged.ECSClusters = append(merged.ECSClusters, cluster)
		}
	}

	return merged
}
