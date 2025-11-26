package main

import (
	"os"
	"testing"
)

func TestCheckSecurityRisks_OpenSSH(t *testing.T) {
	resources := &AWSResources{
		SecurityGroups: []SecurityGroup{
			{
				ID:   "sg-123",
				Name: "test-sg",
				IngressRules: []SGRule{
					{
						Direction: "ingress",
						FromPort:  22,
						ToPort:    22,
						Protocol:  "tcp",
						IPRanges:  []string{"0.0.0.0/0"},
					},
				},
			},
		},
	}

	risks := checkSecurityRisks(resources)

	found := false
	for _, risk := range risks {
		if risk.Type == "Open Security Group" && risk.Severity == "high" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Open Security Group' risk with high severity for open SSH port")
	}
}

func TestCheckSecurityRisks_OpenRDP(t *testing.T) {
	resources := &AWSResources{
		SecurityGroups: []SecurityGroup{
			{
				ID:   "sg-456",
				Name: "rdp-sg",
				IngressRules: []SGRule{
					{
						Direction: "ingress",
						FromPort:  3389,
						ToPort:    3389,
						Protocol:  "tcp",
						IPRanges:  []string{"0.0.0.0/0"},
					},
				},
			},
		},
	}

	risks := checkSecurityRisks(resources)

	found := false
	for _, risk := range risks {
		if risk.Type == "Open Security Group" && risk.ResourceID == "sg-456" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Open Security Group' risk for open RDP port")
	}
}

func TestCheckSecurityRisks_DefaultVPC(t *testing.T) {
	resources := &AWSResources{
		VPCs: []VPC{
			{
				ID:        "vpc-default",
				Name:      "Default VPC",
				IsDefault: true,
			},
		},
	}

	risks := checkSecurityRisks(resources)

	found := false
	for _, risk := range risks {
		if risk.Type == "Default VPC In Use" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Default VPC In Use' risk")
	}
}

func TestCheckSecurityRisks_MissingFlowLogs(t *testing.T) {
	resources := &AWSResources{
		VPCs: []VPC{
			{
				ID:              "vpc-noflow",
				Name:            "No Flow Logs VPC",
				FlowLogsEnabled: false,
			},
		},
	}

	risks := checkSecurityRisks(resources)

	found := false
	for _, risk := range risks {
		if risk.Type == "Missing VPC Flow Logs" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Missing VPC Flow Logs' risk")
	}
}

func TestCheckSecurityRisks_IMDSv2NotEnforced(t *testing.T) {
	resources := &AWSResources{
		Instances: []Instance{
			{
				ID:             "i-12345",
				Name:           "Test Instance",
				IMDSv2Required: false,
				EBSEncrypted:   true,
			},
		},
	}

	risks := checkSecurityRisks(resources)

	found := false
	for _, risk := range risks {
		if risk.Type == "IMDSv2 Not Enforced" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'IMDSv2 Not Enforced' risk")
	}
}

func TestCheckSecurityRisks_UnencryptedEBS(t *testing.T) {
	resources := &AWSResources{
		Instances: []Instance{
			{
				ID:             "i-unencrypted",
				Name:           "Unencrypted Instance",
				IMDSv2Required: true,
				EBSEncrypted:   false,
			},
		},
	}

	risks := checkSecurityRisks(resources)

	found := false
	for _, risk := range risks {
		if risk.Type == "Unencrypted EBS Volume" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Unencrypted EBS Volume' risk")
	}
}

func TestCheckSecurityRisks_PublicRDS(t *testing.T) {
	resources := &AWSResources{
		RDSInstances: []RDSInstance{
			{
				ID:                 "db-public",
				Name:               "Public Database",
				PubliclyAccessible: true,
				Encrypted:          true,
			},
		},
	}

	risks := checkSecurityRisks(resources)

	found := false
	for _, risk := range risks {
		if risk.Type == "Publicly Accessible RDS" && risk.Severity == "critical" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Publicly Accessible RDS' risk with critical severity")
	}
}

func TestCheckSecurityRisks_UnencryptedRDS(t *testing.T) {
	resources := &AWSResources{
		RDSInstances: []RDSInstance{
			{
				ID:                 "db-unencrypted",
				Name:               "Unencrypted Database",
				PubliclyAccessible: false,
				Encrypted:          false,
			},
		},
	}

	risks := checkSecurityRisks(resources)

	found := false
	for _, risk := range risks {
		if risk.Type == "Unencrypted RDS Instance" && risk.Severity == "high" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Unencrypted RDS Instance' risk with high severity")
	}
}

func TestCheckSecurityRisks_OverlyPermissiveIAM(t *testing.T) {
	resources := &AWSResources{
		Instances: []Instance{
			{
				ID:             "i-admin",
				Name:           "Admin Instance",
				IAMRole:        "arn:aws:iam::123456789:role/AdminAccess",
				IMDSv2Required: true,
				EBSEncrypted:   true,
			},
		},
	}

	risks := checkSecurityRisks(resources)

	found := false
	for _, risk := range risks {
		if risk.Type == "Overly Permissive IAM Role" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Overly Permissive IAM Role' risk")
	}
}

func TestGenerateDOTFile(t *testing.T) {
	resources := &AWSResources{
		VPCs: []VPC{
			{ID: "vpc-123", Name: "Test VPC", CIDR: "10.0.0.0/16"},
		},
		Subnets: []Subnet{
			{ID: "subnet-123", VPCID: "vpc-123", Name: "Test Subnet", CIDR: "10.0.1.0/24", AZ: "us-east-1a"},
		},
		Instances: []Instance{
			{ID: "i-123", SubnetID: "subnet-123", Name: "Test Instance", PrivateIP: "10.0.1.10", PublicIP: "N/A"},
		},
	}

	tmpFile := "/tmp/test_network_map.dot"
	defer os.Remove(tmpFile)

	err := generateDOTFile(resources, tmpFile)
	if err != nil {
		t.Fatalf("generateDOTFile failed: %v", err)
	}

	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("DOT file was not created")
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read DOT file: %v", err)
	}

	contentStr := string(content)

	if !containsString(contentStr, "digraph AWS_Network") {
		t.Error("DOT file missing graph declaration")
	}
	// VPC ID "vpc-123" becomes "vpc_vpc-123" with prefix, then dashes become underscores
	if !containsString(contentStr, "vpc_vpc-123") {
		t.Error("DOT file missing VPC node")
	}
}

func TestGenerateDOTFile_EmptyResources(t *testing.T) {
	resources := &AWSResources{}

	tmpFile := "/tmp/test_empty_network_map.dot"
	defer os.Remove(tmpFile)

	err := generateDOTFile(resources, tmpFile)
	if err != nil {
		t.Fatalf("generateDOTFile failed with empty resources: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read DOT file: %v", err)
	}

	if !containsString(string(content), "digraph AWS_Network") {
		t.Error("DOT file missing graph declaration for empty resources")
	}
}

func TestGetPortName(t *testing.T) {
	tests := []struct {
		port     int32
		expected string
	}{
		{22, "SSH"},
		{3389, "RDP"},
		{21, "FTP"},
		{23, "Telnet"},
		{0, "All Ports"},
		{80, "HTTP"},
		{443, "HTTPS"},
		{3306, "MySQL"},
		{5432, "PostgreSQL"},
		{27017, "MongoDB"},
		{6379, "Redis"},
		{8080, "Unknown"},
	}

	for _, tt := range tests {
		result := getPortName(tt.port)
		if result != tt.expected {
			t.Errorf("getPortName(%d) = %s, expected %s", tt.port, result, tt.expected)
		}
	}
}

func TestContainsAdminPattern(t *testing.T) {
	tests := []struct {
		role     string
		expected bool
	}{
		{"arn:aws:iam::123:role/AdminRole", true},
		{"arn:aws:iam::123:role/AdministratorAccess", true},
		{"arn:aws:iam::123:role/PowerUser", true},
		{"arn:aws:iam::123:role/FullAccess", true},
		{"arn:aws:iam::123:role/ReadOnlyAccess", false},
		{"arn:aws:iam::123:role/LambdaExecutionRole", false},
		{"", false},
	}

	for _, tt := range tests {
		result := containsAdminPattern(tt.role)
		if result != tt.expected {
			t.Errorf("containsAdminPattern(%s) = %v, expected %v", tt.role, result, tt.expected)
		}
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"vpc-123", "vpc_123"},
		{"arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/my-lb/abc123", "arn_aws_elasticloadbalancing_us_east_1_123456789_loadbalancer_app_my_lb_abc123"},
		{"simple", "simple"},
		{"with.dots", "with_dots"},
	}

	for _, tt := range tests {
		result := sanitizeID(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeID(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestTruncateLabel(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long label", 10, "this is..."},
		{"exactly10c", 10, "exactly10c"},
		{"", 10, ""},
	}

	for _, tt := range tests {
		result := truncateLabel(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateLabel(%s, %d) = %s, expected %s", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

// Helper function for tests
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringTest(s, substr))
}

func containsSubstringTest(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
