package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// generateMarkdownReport creates a Markdown report of resources and risks
func generateMarkdownReport(resources *AWSResources, risks []Risk, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Header
	fmt.Fprintln(file, "# Cloud NetMapper Report")
	fmt.Fprintf(file, "\nGenerated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(file, "---")

	// Summary
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "## Summary")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "| Resource Type | Count |")
	fmt.Fprintln(file, "|---------------|-------|")
	fmt.Fprintf(file, "| VPCs | %d |\n", len(resources.VPCs))
	fmt.Fprintf(file, "| Subnets | %d |\n", len(resources.Subnets))
	fmt.Fprintf(file, "| EC2 Instances | %d |\n", len(resources.Instances))
	fmt.Fprintf(file, "| Security Groups | %d |\n", len(resources.SecurityGroups))
	fmt.Fprintf(file, "| Load Balancers | %d |\n", len(resources.LoadBalancers))
	fmt.Fprintf(file, "| RDS Instances | %d |\n", len(resources.RDSInstances))
	fmt.Fprintf(file, "| NAT Gateways | %d |\n", len(resources.NATGateways))
	fmt.Fprintf(file, "| Internet Gateways | %d |\n", len(resources.InternetGateways))
	fmt.Fprintf(file, "| Route Tables | %d |\n", len(resources.RouteTables))
	fmt.Fprintf(file, "| VPC Peerings | %d |\n", len(resources.VPCPeerings))
	fmt.Fprintf(file, "| Lambda Functions | %d |\n", len(resources.LambdaFunctions))
	fmt.Fprintf(file, "| EKS Clusters | %d |\n", len(resources.EKSClusters))
	fmt.Fprintf(file, "| ECS Clusters | %d |\n", len(resources.ECSClusters))

	// Security Risks
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "## Security Risks")
	fmt.Fprintln(file, "")
	if len(risks) == 0 {
		fmt.Fprintln(file, "No security risks detected.")
		fmt.Fprintln(file, "")
	} else {
		// Count by severity
		critical, high, medium, low := 0, 0, 0, 0
		for _, risk := range risks {
			switch risk.Severity {
			case "critical":
				critical++
			case "high":
				high++
			case "medium":
				medium++
			case "low":
				low++
			}
		}
		fmt.Fprintf(file, "**Total Risks:** %d (Critical: %d, High: %d, Medium: %d, Low: %d)\n\n", len(risks), critical, high, medium, low)

		fmt.Fprintln(file, "| Severity | Type | Resource | Details | Remediation |")
		fmt.Fprintln(file, "|----------|------|----------|---------|-------------|")
		for _, risk := range risks {
			fmt.Fprintf(file, "| %s | %s | %s | %s | %s |\n",
				strings.ToUpper(risk.Severity),
				risk.Type,
				risk.Resource,
				risk.Details,
				risk.Remediation)
		}
	}

	// VPCs
	if len(resources.VPCs) > 0 {
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "## VPCs")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Name | ID | CIDR | Default | Flow Logs |")
		fmt.Fprintln(file, "|------|-----|------|---------|-----------|")
		for _, vpc := range resources.VPCs {
			fmt.Fprintf(file, "| %s | %s | %s | %v | %v |\n",
				vpc.Name, vpc.ID, vpc.CIDR, vpc.IsDefault, vpc.FlowLogsEnabled)
		}
	}

	// Subnets
	if len(resources.Subnets) > 0 {
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "## Subnets")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Name | ID | VPC ID | CIDR | AZ | Public |")
		fmt.Fprintln(file, "|------|-----|--------|------|-----|--------|")
		for _, subnet := range resources.Subnets {
			fmt.Fprintf(file, "| %s | %s | %s | %s | %s | %v |\n",
				subnet.Name, subnet.ID, subnet.VPCID, subnet.CIDR, subnet.AZ, subnet.IsPublic)
		}
	}

	// EC2 Instances
	if len(resources.Instances) > 0 {
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "## EC2 Instances")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Name | ID | Type | Private IP | Public IP | IMDSv2 |")
		fmt.Fprintln(file, "|------|-----|------|------------|-----------|--------|")
		for _, inst := range resources.Instances {
			fmt.Fprintf(file, "| %s | %s | %s | %s | %s | %v |\n",
				inst.Name, inst.ID, inst.InstanceType, inst.PrivateIP, inst.PublicIP, inst.IMDSv2Required)
		}
	}

	// RDS Instances
	if len(resources.RDSInstances) > 0 {
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "## RDS Instances")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Name | Engine | Class | Encrypted | Public |")
		fmt.Fprintln(file, "|------|--------|-------|-----------|--------|")
		for _, rds := range resources.RDSInstances {
			fmt.Fprintf(file, "| %s | %s | %s | %v | %v |\n",
				rds.Name, rds.Engine, rds.InstanceClass, rds.Encrypted, rds.PubliclyAccessible)
		}
	}

	// Load Balancers
	if len(resources.LoadBalancers) > 0 {
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "## Load Balancers")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Name | Type | Scheme | VPC ID |")
		fmt.Fprintln(file, "|------|------|--------|--------|")
		for _, lb := range resources.LoadBalancers {
			fmt.Fprintf(file, "| %s | %s | %s | %s |\n",
				lb.Name, lb.Type, lb.Scheme, lb.VPCID)
		}
	}

	// Lambda Functions
	if len(resources.LambdaFunctions) > 0 {
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "## Lambda Functions")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Name | Runtime | VPC ID |")
		fmt.Fprintln(file, "|------|---------|--------|")
		for _, fn := range resources.LambdaFunctions {
			vpcID := fn.VPCID
			if vpcID == "" {
				vpcID = "N/A"
			}
			fmt.Fprintf(file, "| %s | %s | %s |\n", fn.Name, fn.Runtime, vpcID)
		}
	}

	// EKS Clusters
	if len(resources.EKSClusters) > 0 {
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "## EKS Clusters")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Name | Status | VPC ID |")
		fmt.Fprintln(file, "|------|--------|--------|")
		for _, cluster := range resources.EKSClusters {
			fmt.Fprintf(file, "| %s | %s | %s |\n", cluster.Name, cluster.Status, cluster.VPCID)
		}
	}

	// ECS Clusters
	if len(resources.ECSClusters) > 0 {
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "## ECS Clusters")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Name | Status | Services | Tasks |")
		fmt.Fprintln(file, "|------|--------|----------|-------|")
		for _, cluster := range resources.ECSClusters {
			fmt.Fprintf(file, "| %s | %s | %d | %d |\n",
				cluster.Name, cluster.Status, cluster.ServiceCount, cluster.TaskCount)
		}
	}

	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "---")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "*Report generated by Cloud NetMapper*")

	return nil
}

// generateCSVReport creates a CSV export of all resources
func generateCSVReport(resources *AWSResources, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Type", "Name", "ID", "VPC ID", "Details", "Tags"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write VPCs
	for _, vpc := range resources.VPCs {
		details := fmt.Sprintf("CIDR: %s, Default: %v, FlowLogs: %v", vpc.CIDR, vpc.IsDefault, vpc.FlowLogsEnabled)
		row := []string{"VPC", vpc.Name, vpc.ID, vpc.ID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write Subnets
	for _, subnet := range resources.Subnets {
		details := fmt.Sprintf("CIDR: %s, AZ: %s, Public: %v", subnet.CIDR, subnet.AZ, subnet.IsPublic)
		row := []string{"Subnet", subnet.Name, subnet.ID, subnet.VPCID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write Instances
	for _, inst := range resources.Instances {
		details := fmt.Sprintf("Type: %s, Private: %s, Public: %s", inst.InstanceType, inst.PrivateIP, inst.PublicIP)
		row := []string{"EC2", inst.Name, inst.ID, inst.VPCID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write Security Groups
	for _, sg := range resources.SecurityGroups {
		details := fmt.Sprintf("Ingress Rules: %d, Egress Rules: %d", len(sg.IngressRules), len(sg.EgressRules))
		row := []string{"SecurityGroup", sg.Name, sg.ID, sg.VPCID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write Load Balancers
	for _, lb := range resources.LoadBalancers {
		details := fmt.Sprintf("Type: %s, Scheme: %s", lb.Type, lb.Scheme)
		row := []string{"LoadBalancer", lb.Name, lb.ARN, lb.VPCID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write RDS Instances
	for _, rds := range resources.RDSInstances {
		details := fmt.Sprintf("Engine: %s, Class: %s, Encrypted: %v, Public: %v",
			rds.Engine, rds.InstanceClass, rds.Encrypted, rds.PubliclyAccessible)
		row := []string{"RDS", rds.Name, rds.ID, rds.VPCID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write NAT Gateways
	for _, nat := range resources.NATGateways {
		details := fmt.Sprintf("State: %s, Public IP: %s", nat.State, nat.PublicIP)
		row := []string{"NATGateway", nat.Name, nat.ID, nat.VPCID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write Internet Gateways
	for _, igw := range resources.InternetGateways {
		details := fmt.Sprintf("Attached VPCs: %s", strings.Join(igw.VPCIDs, ", "))
		vpcID := ""
		if len(igw.VPCIDs) > 0 {
			vpcID = igw.VPCIDs[0]
		}
		row := []string{"InternetGateway", igw.Name, igw.ID, vpcID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write Lambda Functions
	for _, fn := range resources.LambdaFunctions {
		details := fmt.Sprintf("Runtime: %s", fn.Runtime)
		row := []string{"Lambda", fn.Name, fn.ARN, fn.VPCID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write EKS Clusters
	for _, cluster := range resources.EKSClusters {
		details := fmt.Sprintf("Status: %s", cluster.Status)
		row := []string{"EKS", cluster.Name, cluster.ARN, cluster.VPCID, details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write ECS Clusters
	for _, cluster := range resources.ECSClusters {
		details := fmt.Sprintf("Status: %s, Services: %d, Tasks: %d",
			cluster.Status, cluster.ServiceCount, cluster.TaskCount)
		row := []string{"ECS", cluster.Name, cluster.ARN, "", details, ""}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// SARIF types for security report
type SARIFReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []SARIFRule `json:"rules"`
}

type SARIFRule struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	ShortDescription SARIFDescription `json:"shortDescription"`
	FullDescription  SARIFDescription `json:"fullDescription"`
	Help             SARIFDescription `json:"help"`
	DefaultConfig    SARIFConfig      `json:"defaultConfiguration"`
}

type SARIFDescription struct {
	Text string `json:"text"`
}

type SARIFConfig struct {
	Level string `json:"level"`
}

type SARIFResult struct {
	RuleID    string           `json:"ruleId"`
	Level     string           `json:"level"`
	Message   SARIFDescription `json:"message"`
	Locations []SARIFLocation  `json:"locations"`
}

type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
}

type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

// generateSARIFReport creates a SARIF format security report
func generateSARIFReport(risks []Risk, filename string) error {
	// Build rules from unique risk types
	ruleMap := make(map[string]SARIFRule)
	for _, risk := range risks {
		ruleID := strings.ReplaceAll(risk.Type, " ", "-")
		if _, exists := ruleMap[ruleID]; !exists {
			ruleMap[ruleID] = SARIFRule{
				ID:               ruleID,
				Name:             risk.Type,
				ShortDescription: SARIFDescription{Text: risk.Type},
				FullDescription:  SARIFDescription{Text: risk.Details},
				Help:             SARIFDescription{Text: risk.Remediation},
				DefaultConfig:    SARIFConfig{Level: mapSeverityToSARIF(risk.Severity)},
			}
		}
	}

	var rules []SARIFRule
	for _, rule := range ruleMap {
		rules = append(rules, rule)
	}

	// Build results
	var results []SARIFResult
	for _, risk := range risks {
		ruleID := strings.ReplaceAll(risk.Type, " ", "-")
		result := SARIFResult{
			RuleID:  ruleID,
			Level:   mapSeverityToSARIF(risk.Severity),
			Message: SARIFDescription{Text: fmt.Sprintf("%s: %s", risk.Details, risk.Remediation)},
			Locations: []SARIFLocation{
				{
					PhysicalLocation: SARIFPhysicalLocation{
						ArtifactLocation: SARIFArtifactLocation{
							URI: fmt.Sprintf("aws://%s/%s", risk.Resource, risk.ResourceID),
						},
					},
				},
			},
		}
		results = append(results, result)
	}

	report := SARIFReport{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "cloud-netmapper",
						Version:        "1.0.0",
						InformationURI: "https://github.com/msaadshabir/cloud-netmapper",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func mapSeverityToSARIF(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "error"
	case "high":
		return "error"
	case "medium":
		return "warning"
	case "low":
		return "note"
	default:
		return "note"
	}
}
