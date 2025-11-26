package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ScanSnapshot represents a point-in-time scan of AWS resources
type ScanSnapshot struct {
	Timestamp time.Time     `json:"timestamp"`
	Region    string        `json:"region"`
	Resources *AWSResources `json:"resources"`
}

// ResourceChange represents a detected change in the infrastructure
type ResourceChange struct {
	ChangeType   string `json:"change_type"`   // "added", "removed", "modified"
	ResourceType string `json:"resource_type"` // "VPC", "Subnet", "Instance", etc.
	ResourceID   string `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	Details      string `json:"details"`
}

// DiffReport contains all detected changes between two scans
type DiffReport struct {
	PreviousScan time.Time        `json:"previous_scan"`
	CurrentScan  time.Time        `json:"current_scan"`
	Changes      []ResourceChange `json:"changes"`
	Summary      DiffSummary      `json:"summary"`
}

// DiffSummary provides counts of changes by type
type DiffSummary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
	Total    int `json:"total"`
}

// SaveSnapshot saves the current scan to a JSON file for future comparison
func SaveSnapshot(resources *AWSResources, region, outputDir string) error {
	snapshot := ScanSnapshot{
		Timestamp: time.Now(),
		Region:    region,
		Resources: resources,
	}

	// Create snapshots directory if it doesn't exist
	snapshotDir := filepath.Join(outputDir, "snapshots")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	// Save as timestamped file
	filename := filepath.Join(snapshotDir, fmt.Sprintf("scan_%s_%s.json",
		region,
		time.Now().Format("20060102_150405")))

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	// Also save as "latest" for easy comparison
	latestFile := filepath.Join(snapshotDir, fmt.Sprintf("scan_%s_latest.json", region))
	if err := os.WriteFile(latestFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write latest snapshot: %w", err)
	}

	return nil
}

// LoadPreviousSnapshot loads the most recent snapshot for comparison
func LoadPreviousSnapshot(region, outputDir string) (*ScanSnapshot, error) {
	snapshotDir := filepath.Join(outputDir, "snapshots")
	latestFile := filepath.Join(snapshotDir, fmt.Sprintf("scan_%s_latest.json", region))

	data, err := os.ReadFile(latestFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No previous snapshot exists
		}
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	var snapshot ScanSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	return &snapshot, nil
}

// DetectChanges compares current resources with a previous snapshot
func DetectChanges(current, previous *AWSResources) *DiffReport {
	report := &DiffReport{
		CurrentScan: time.Now(),
		Changes:     []ResourceChange{},
	}

	if previous == nil {
		return report
	}

	// Compare VPCs
	report.Changes = append(report.Changes, compareVPCs(current.VPCs, previous.VPCs)...)

	// Compare Subnets
	report.Changes = append(report.Changes, compareSubnets(current.Subnets, previous.Subnets)...)

	// Compare Instances
	report.Changes = append(report.Changes, compareInstances(current.Instances, previous.Instances)...)

	// Compare Security Groups
	report.Changes = append(report.Changes, compareSecurityGroups(current.SecurityGroups, previous.SecurityGroups)...)

	// Compare Load Balancers
	report.Changes = append(report.Changes, compareLoadBalancers(current.LoadBalancers, previous.LoadBalancers)...)

	// Compare RDS Instances
	report.Changes = append(report.Changes, compareRDSInstances(current.RDSInstances, previous.RDSInstances)...)

	// Compare NAT Gateways
	report.Changes = append(report.Changes, compareNATGateways(current.NATGateways, previous.NATGateways)...)

	// Compare Internet Gateways
	report.Changes = append(report.Changes, compareInternetGateways(current.InternetGateways, previous.InternetGateways)...)

	// Compare Lambda Functions
	report.Changes = append(report.Changes, compareLambdaFunctions(current.LambdaFunctions, previous.LambdaFunctions)...)

	// Compare EKS Clusters
	report.Changes = append(report.Changes, compareEKSClusters(current.EKSClusters, previous.EKSClusters)...)

	// Compare ECS Clusters
	report.Changes = append(report.Changes, compareECSClusters(current.ECSClusters, previous.ECSClusters)...)

	// Calculate summary
	for _, change := range report.Changes {
		switch change.ChangeType {
		case "added":
			report.Summary.Added++
		case "removed":
			report.Summary.Removed++
		case "modified":
			report.Summary.Modified++
		}
	}
	report.Summary.Total = len(report.Changes)

	return report
}

func compareVPCs(current, previous []VPC) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]VPC)
	previousMap := make(map[string]VPC)

	for _, vpc := range current {
		currentMap[vpc.ID] = vpc
	}
	for _, vpc := range previous {
		previousMap[vpc.ID] = vpc
	}

	// Check for added and modified VPCs
	for id, curr := range currentMap {
		if prev, exists := previousMap[id]; exists {
			// Check for modifications
			if curr.Name != prev.Name || curr.CIDR != prev.CIDR || curr.FlowLogsEnabled != prev.FlowLogsEnabled {
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "VPC",
					ResourceID:   id,
					ResourceName: curr.Name,
					Details:      "VPC configuration changed",
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "VPC",
				ResourceID:   id,
				ResourceName: curr.Name,
				Details:      fmt.Sprintf("New VPC created with CIDR %s", curr.CIDR),
			})
		}
	}

	// Check for removed VPCs
	for id, prev := range previousMap {
		if _, exists := currentMap[id]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "VPC",
				ResourceID:   id,
				ResourceName: prev.Name,
				Details:      fmt.Sprintf("VPC deleted (was CIDR %s)", prev.CIDR),
			})
		}
	}

	return changes
}

func compareSubnets(current, previous []Subnet) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]Subnet)
	previousMap := make(map[string]Subnet)

	for _, subnet := range current {
		currentMap[subnet.ID] = subnet
	}
	for _, subnet := range previous {
		previousMap[subnet.ID] = subnet
	}

	for id, curr := range currentMap {
		if prev, exists := previousMap[id]; exists {
			if curr.Name != prev.Name || curr.CIDR != prev.CIDR || curr.IsPublic != prev.IsPublic {
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "Subnet",
					ResourceID:   id,
					ResourceName: curr.Name,
					Details:      "Subnet configuration changed",
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "Subnet",
				ResourceID:   id,
				ResourceName: curr.Name,
				Details:      fmt.Sprintf("New subnet in %s with CIDR %s", curr.AZ, curr.CIDR),
			})
		}
	}

	for id, prev := range previousMap {
		if _, exists := currentMap[id]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "Subnet",
				ResourceID:   id,
				ResourceName: prev.Name,
				Details:      fmt.Sprintf("Subnet deleted from %s", prev.AZ),
			})
		}
	}

	return changes
}

func compareInstances(current, previous []Instance) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]Instance)
	previousMap := make(map[string]Instance)

	for _, inst := range current {
		currentMap[inst.ID] = inst
	}
	for _, inst := range previous {
		previousMap[inst.ID] = inst
	}

	for id, curr := range currentMap {
		if prev, exists := previousMap[id]; exists {
			if curr.InstanceType != prev.InstanceType ||
				curr.PublicIP != prev.PublicIP || curr.IAMRole != prev.IAMRole ||
				curr.SubnetID != prev.SubnetID {
				details := []string{}
				if curr.InstanceType != prev.InstanceType {
					details = append(details, fmt.Sprintf("type: %s -> %s", prev.InstanceType, curr.InstanceType))
				}
				if curr.PublicIP != prev.PublicIP {
					details = append(details, fmt.Sprintf("public IP: %s -> %s", prev.PublicIP, curr.PublicIP))
				}
				if curr.SubnetID != prev.SubnetID {
					details = append(details, fmt.Sprintf("subnet: %s -> %s", prev.SubnetID, curr.SubnetID))
				}
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "EC2 Instance",
					ResourceID:   id,
					ResourceName: curr.Name,
					Details:      fmt.Sprintf("Changes: %v", details),
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "EC2 Instance",
				ResourceID:   id,
				ResourceName: curr.Name,
				Details:      fmt.Sprintf("New %s instance", curr.InstanceType),
			})
		}
	}

	for id, prev := range previousMap {
		if _, exists := currentMap[id]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "EC2 Instance",
				ResourceID:   id,
				ResourceName: prev.Name,
				Details:      fmt.Sprintf("%s instance terminated", prev.InstanceType),
			})
		}
	}

	return changes
}

func compareSecurityGroups(current, previous []SecurityGroup) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]SecurityGroup)
	previousMap := make(map[string]SecurityGroup)

	for _, sg := range current {
		currentMap[sg.ID] = sg
	}
	for _, sg := range previous {
		previousMap[sg.ID] = sg
	}

	for id, curr := range currentMap {
		if prev, exists := previousMap[id]; exists {
			// Compare rule counts
			if len(curr.IngressRules) != len(prev.IngressRules) || len(curr.EgressRules) != len(prev.EgressRules) {
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "Security Group",
					ResourceID:   id,
					ResourceName: curr.Name,
					Details: fmt.Sprintf("Rules changed: ingress %d->%d, egress %d->%d",
						len(prev.IngressRules), len(curr.IngressRules),
						len(prev.EgressRules), len(curr.EgressRules)),
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "Security Group",
				ResourceID:   id,
				ResourceName: curr.Name,
				Details:      fmt.Sprintf("New security group with %d ingress and %d egress rules", len(curr.IngressRules), len(curr.EgressRules)),
			})
		}
	}

	for id, prev := range previousMap {
		if _, exists := currentMap[id]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "Security Group",
				ResourceID:   id,
				ResourceName: prev.Name,
				Details:      "Security group deleted",
			})
		}
	}

	return changes
}

func compareLoadBalancers(current, previous []LoadBalancer) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]LoadBalancer)
	previousMap := make(map[string]LoadBalancer)

	for _, lb := range current {
		currentMap[lb.ARN] = lb
	}
	for _, lb := range previous {
		previousMap[lb.ARN] = lb
	}

	for arn, curr := range currentMap {
		if prev, exists := previousMap[arn]; exists {
			if curr.Scheme != prev.Scheme {
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "Load Balancer",
					ResourceID:   arn,
					ResourceName: curr.Name,
					Details:      fmt.Sprintf("Scheme changed: %s -> %s", prev.Scheme, curr.Scheme),
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "Load Balancer",
				ResourceID:   arn,
				ResourceName: curr.Name,
				Details:      fmt.Sprintf("New %s load balancer (%s)", curr.Type, curr.Scheme),
			})
		}
	}

	for arn, prev := range previousMap {
		if _, exists := currentMap[arn]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "Load Balancer",
				ResourceID:   arn,
				ResourceName: prev.Name,
				Details:      "Load balancer deleted",
			})
		}
	}

	return changes
}

func compareRDSInstances(current, previous []RDSInstance) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]RDSInstance)
	previousMap := make(map[string]RDSInstance)

	for _, rds := range current {
		currentMap[rds.ID] = rds
	}
	for _, rds := range previous {
		previousMap[rds.ID] = rds
	}

	for id, curr := range currentMap {
		if prev, exists := previousMap[id]; exists {
			if curr.InstanceClass != prev.InstanceClass || curr.PubliclyAccessible != prev.PubliclyAccessible {
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "RDS Instance",
					ResourceID:   id,
					ResourceName: curr.Name,
					Details:      "RDS configuration changed",
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "RDS Instance",
				ResourceID:   id,
				ResourceName: curr.Name,
				Details:      fmt.Sprintf("New %s RDS instance (%s)", curr.Engine, curr.InstanceClass),
			})
		}
	}

	for id, prev := range previousMap {
		if _, exists := currentMap[id]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "RDS Instance",
				ResourceID:   id,
				ResourceName: prev.Name,
				Details:      fmt.Sprintf("%s RDS instance deleted", prev.Engine),
			})
		}
	}

	return changes
}

func compareNATGateways(current, previous []NATGateway) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]NATGateway)
	previousMap := make(map[string]NATGateway)

	for _, nat := range current {
		currentMap[nat.ID] = nat
	}
	for _, nat := range previous {
		previousMap[nat.ID] = nat
	}

	for id, curr := range currentMap {
		if prev, exists := previousMap[id]; exists {
			if curr.State != prev.State {
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "NAT Gateway",
					ResourceID:   id,
					ResourceName: curr.Name,
					Details:      fmt.Sprintf("State changed: %s -> %s", prev.State, curr.State),
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "NAT Gateway",
				ResourceID:   id,
				ResourceName: curr.Name,
				Details:      fmt.Sprintf("New NAT Gateway with public IP %s", curr.PublicIP),
			})
		}
	}

	for id, prev := range previousMap {
		if _, exists := currentMap[id]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "NAT Gateway",
				ResourceID:   id,
				ResourceName: prev.Name,
				Details:      "NAT Gateway deleted",
			})
		}
	}

	return changes
}

func compareInternetGateways(current, previous []InternetGateway) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]InternetGateway)
	previousMap := make(map[string]InternetGateway)

	for _, igw := range current {
		currentMap[igw.ID] = igw
	}
	for _, igw := range previous {
		previousMap[igw.ID] = igw
	}

	for id, curr := range currentMap {
		if _, exists := previousMap[id]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "Internet Gateway",
				ResourceID:   id,
				ResourceName: curr.Name,
				Details:      "New Internet Gateway created",
			})
		}
	}

	for id, prev := range previousMap {
		if _, exists := currentMap[id]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "Internet Gateway",
				ResourceID:   id,
				ResourceName: prev.Name,
				Details:      "Internet Gateway deleted",
			})
		}
	}

	return changes
}

func compareLambdaFunctions(current, previous []LambdaFunction) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]LambdaFunction)
	previousMap := make(map[string]LambdaFunction)

	for _, fn := range current {
		currentMap[fn.Name] = fn
	}
	for _, fn := range previous {
		previousMap[fn.Name] = fn
	}

	for name, curr := range currentMap {
		if prev, exists := previousMap[name]; exists {
			if curr.Runtime != prev.Runtime || curr.VPCID != prev.VPCID {
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "Lambda Function",
					ResourceID:   curr.ARN,
					ResourceName: name,
					Details:      fmt.Sprintf("Runtime: %s, VPC: %s", curr.Runtime, curr.VPCID),
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "Lambda Function",
				ResourceID:   curr.ARN,
				ResourceName: name,
				Details:      fmt.Sprintf("New Lambda function with runtime %s", curr.Runtime),
			})
		}
	}

	for name, prev := range previousMap {
		if _, exists := currentMap[name]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "Lambda Function",
				ResourceID:   prev.ARN,
				ResourceName: name,
				Details:      "Lambda function deleted",
			})
		}
	}

	return changes
}

func compareEKSClusters(current, previous []EKSCluster) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]EKSCluster)
	previousMap := make(map[string]EKSCluster)

	for _, cluster := range current {
		currentMap[cluster.Name] = cluster
	}
	for _, cluster := range previous {
		previousMap[cluster.Name] = cluster
	}

	for name, curr := range currentMap {
		if prev, exists := previousMap[name]; exists {
			if curr.Status != prev.Status || curr.VPCID != prev.VPCID {
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "EKS Cluster",
					ResourceID:   curr.ARN,
					ResourceName: name,
					Details:      fmt.Sprintf("Status: %s, VPC: %s", curr.Status, curr.VPCID),
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "EKS Cluster",
				ResourceID:   curr.ARN,
				ResourceName: name,
				Details:      fmt.Sprintf("New EKS cluster (status: %s)", curr.Status),
			})
		}
	}

	for name, prev := range previousMap {
		if _, exists := currentMap[name]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "EKS Cluster",
				ResourceID:   prev.ARN,
				ResourceName: name,
				Details:      "EKS cluster deleted",
			})
		}
	}

	return changes
}

func compareECSClusters(current, previous []ECSCluster) []ResourceChange {
	var changes []ResourceChange
	currentMap := make(map[string]ECSCluster)
	previousMap := make(map[string]ECSCluster)

	for _, cluster := range current {
		currentMap[cluster.Name] = cluster
	}
	for _, cluster := range previous {
		previousMap[cluster.Name] = cluster
	}

	for name, curr := range currentMap {
		if prev, exists := previousMap[name]; exists {
			if curr.ServiceCount != prev.ServiceCount || curr.TaskCount != prev.TaskCount {
				changes = append(changes, ResourceChange{
					ChangeType:   "modified",
					ResourceType: "ECS Cluster",
					ResourceID:   curr.ARN,
					ResourceName: name,
					Details: fmt.Sprintf("Services: %d->%d, Tasks: %d->%d",
						prev.ServiceCount, curr.ServiceCount,
						prev.TaskCount, curr.TaskCount),
				})
			}
		} else {
			changes = append(changes, ResourceChange{
				ChangeType:   "added",
				ResourceType: "ECS Cluster",
				ResourceID:   curr.ARN,
				ResourceName: name,
				Details:      fmt.Sprintf("New ECS cluster with %d services", curr.ServiceCount),
			})
		}
	}

	for name, prev := range previousMap {
		if _, exists := currentMap[name]; !exists {
			changes = append(changes, ResourceChange{
				ChangeType:   "removed",
				ResourceType: "ECS Cluster",
				ResourceID:   prev.ARN,
				ResourceName: name,
				Details:      "ECS cluster deleted",
			})
		}
	}

	return changes
}

// GenerateDiffReport creates a human-readable diff report
func GenerateDiffReport(report *DiffReport, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "# Infrastructure Change Report")
	fmt.Fprintln(file, "")
	fmt.Fprintf(file, "Previous scan: %s\n", report.PreviousScan.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "Current scan: %s\n", report.CurrentScan.Format("2006-01-02 15:04:05"))
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "## Summary")
	fmt.Fprintln(file, "")
	fmt.Fprintf(file, "- Added: %d\n", report.Summary.Added)
	fmt.Fprintf(file, "- Removed: %d\n", report.Summary.Removed)
	fmt.Fprintf(file, "- Modified: %d\n", report.Summary.Modified)
	fmt.Fprintf(file, "- Total changes: %d\n", report.Summary.Total)
	fmt.Fprintln(file, "")

	if report.Summary.Total == 0 {
		fmt.Fprintln(file, "No changes detected.")
		return nil
	}

	// Group changes by type
	added := []ResourceChange{}
	removed := []ResourceChange{}
	modified := []ResourceChange{}

	for _, change := range report.Changes {
		switch change.ChangeType {
		case "added":
			added = append(added, change)
		case "removed":
			removed = append(removed, change)
		case "modified":
			modified = append(modified, change)
		}
	}

	if len(added) > 0 {
		fmt.Fprintln(file, "## Added Resources")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Type | Name | ID | Details |")
		fmt.Fprintln(file, "|------|------|-----|---------|")
		for _, change := range added {
			fmt.Fprintf(file, "| %s | %s | %s | %s |\n",
				change.ResourceType, change.ResourceName, change.ResourceID, change.Details)
		}
		fmt.Fprintln(file, "")
	}

	if len(removed) > 0 {
		fmt.Fprintln(file, "## Removed Resources")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Type | Name | ID | Details |")
		fmt.Fprintln(file, "|------|------|-----|---------|")
		for _, change := range removed {
			fmt.Fprintf(file, "| %s | %s | %s | %s |\n",
				change.ResourceType, change.ResourceName, change.ResourceID, change.Details)
		}
		fmt.Fprintln(file, "")
	}

	if len(modified) > 0 {
		fmt.Fprintln(file, "## Modified Resources")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "| Type | Name | ID | Details |")
		fmt.Fprintln(file, "|------|------|-----|---------|")
		for _, change := range modified {
			fmt.Fprintf(file, "| %s | %s | %s | %s |\n",
				change.ResourceType, change.ResourceName, change.ResourceID, change.Details)
		}
		fmt.Fprintln(file, "")
	}

	fmt.Fprintln(file, "---")
	fmt.Fprintln(file, "*Report generated by Cloud NetMapper*")

	return nil
}
