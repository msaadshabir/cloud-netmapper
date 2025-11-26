package types

// VPC represents an AWS VPC
type VPC struct {
	ID              string
	CIDR            string
	Name            string
	IsDefault       bool
	FlowLogsEnabled bool
}

// Subnet represents an AWS Subnet
type Subnet struct {
	ID       string
	VPCID    string
	CIDR     string
	AZ       string
	Name     string
	IsPublic bool
}

// Instance represents an EC2 instance
type Instance struct {
	ID             string
	VPCID          string
	SubnetID       string
	PrivateIP      string
	PublicIP       string
	SGIDs          []string
	Name           string
	InstanceType   string
	IMDSv2Required bool
	IAMRole        string
	EBSEncrypted   bool
}

// SecurityGroup represents an AWS Security Group
type SecurityGroup struct {
	ID           string
	VPCID        string
	Name         string
	Description  string
	IngressRules []SGRule
	EgressRules  []SGRule
}

// SGRule represents a Security Group rule
type SGRule struct {
	Direction   string // "ingress" or "egress"
	FromPort    int32
	ToPort      int32
	Protocol    string
	IPRanges    []string
	Description string
}

// LoadBalancer represents an ELB/ALB/NLB
type LoadBalancer struct {
	ARN       string
	Name      string
	VPCID     string
	Scheme    string
	Type      string
	SubnetIDs []string
}

// RDSInstance represents an RDS database instance
type RDSInstance struct {
	ID                 string
	VPCID              string
	Engine             string
	EngineVersion      string
	InstanceClass      string
	PubliclyAccessible bool
	Encrypted          bool
	SubnetGroupName    string
	SecurityGroupIDs   []string
	Name               string
}

// NATGateway represents a NAT Gateway
type NATGateway struct {
	ID        string
	VPCID     string
	SubnetID  string
	PublicIP  string
	PrivateIP string
	State     string
	Name      string
}

// InternetGateway represents an Internet Gateway
type InternetGateway struct {
	ID     string
	VPCIDs []string
	Name   string
}

// RouteTable represents a Route Table
type RouteTable struct {
	ID     string
	VPCID  string
	Name   string
	Routes []Route
	IsMain bool
}

// Route represents a route in a route table
type Route struct {
	DestinationCIDR string
	GatewayID       string
	NATGatewayID    string
	InstanceID      string
	State           string
}

// VPCPeering represents a VPC Peering connection
type VPCPeering struct {
	ID             string
	RequesterVPCID string
	AccepterVPCID  string
	Status         string
	Name           string
}

// LambdaFunction represents a Lambda function with VPC config
type LambdaFunction struct {
	Name             string
	ARN              string
	Runtime          string
	VPCID            string
	SubnetIDs        []string
	SecurityGroupIDs []string
}

// EKSCluster represents an EKS cluster
type EKSCluster struct {
	Name            string
	ARN             string
	VPCID           string
	SubnetIDs       []string
	SecurityGroupID string
	Status          string
}

// ECSCluster represents an ECS cluster
type ECSCluster struct {
	Name         string
	ARN          string
	Status       string
	ServiceCount int
	TaskCount    int
}

// CloudResources is the common interface for cloud resources
type CloudResources struct {
	Provider         string
	Region           string
	VPCs             []VPC
	Subnets          []Subnet
	Instances        []Instance
	SecurityGroups   []SecurityGroup
	LoadBalancers    []LoadBalancer
	RDSInstances     []RDSInstance
	NATGateways      []NATGateway
	InternetGateways []InternetGateway
	RouteTables      []RouteTable
	VPCPeerings      []VPCPeering
	LambdaFunctions  []LambdaFunction
	EKSClusters      []EKSCluster
	ECSClusters      []ECSCluster
}

// Risk represents a security risk
type Risk struct {
	Type        string
	Resource    string
	ResourceID  string
	Details     string
	Severity    string // "low", "medium", "high", "critical"
	Remediation string
}

// ScanResult contains the complete scan result
type ScanResult struct {
	Timestamp string
	Regions   []string
	Resources map[string]*CloudResources // keyed by region
	Risks     []Risk
	Summary   ScanSummary
}

// ScanSummary contains summary statistics
type ScanSummary struct {
	TotalVPCs             int
	TotalSubnets          int
	TotalInstances        int
	TotalSecurityGroups   int
	TotalLoadBalancers    int
	TotalRDSInstances     int
	TotalNATGateways      int
	TotalInternetGateways int
	TotalRouteTables      int
	TotalVPCPeerings      int
	TotalLambdaFunctions  int
	TotalEKSClusters      int
	TotalECSClusters      int
	CriticalRisks         int
	HighRisks             int
	MediumRisks           int
	LowRisks              int
}
