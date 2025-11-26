package providers

import (
	"context"
	"fmt"

	"cloud-netmapper/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// AWSProvider implements the CloudProvider interface for AWS
type AWSProvider struct {
	config aws.Config
}

// NewAWSProvider creates a new AWS provider instance
func NewAWSProvider() (*AWSProvider, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return &AWSProvider{config: cfg}, nil
}

// NewAWSProviderForRegion creates an AWS provider configured for a specific region
func NewAWSProviderForRegion(region string) (*AWSProvider, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
	}
	return &AWSProvider{config: cfg}, nil
}

// Name returns the provider name
func (p *AWSProvider) Name() string {
	return "AWS"
}

// ListRegions returns all available AWS regions
func (p *AWSProvider) ListRegions() ([]string, error) {
	ec2Client := ec2.NewFromConfig(p.config)
	resp, err := ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	var regions []string
	for _, region := range resp.Regions {
		if region.RegionName != nil {
			regions = append(regions, *region.RegionName)
		}
	}
	return regions, nil
}

// GetResources retrieves all resources from the specified region
func (p *AWSProvider) GetResources(region string) (*types.CloudResources, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load config for region %s: %w", region, err)
	}

	resources := &types.CloudResources{
		Provider: "AWS",
		Region:   region,
	}

	ec2Client := ec2.NewFromConfig(cfg)
	rdsClient := rds.NewFromConfig(cfg)
	elbClient := elbv2.NewFromConfig(cfg)
	lambdaClient := lambda.NewFromConfig(cfg)
	eksClient := eks.NewFromConfig(cfg)
	ecsClient := ecs.NewFromConfig(cfg)

	// Collect VPCs
	vpcs, err := p.collectVPCs(ec2Client)
	if err != nil {
		return nil, fmt.Errorf("failed to collect VPCs: %w", err)
	}
	resources.VPCs = vpcs

	// Collect Subnets
	subnets, err := p.collectSubnets(ec2Client)
	if err != nil {
		return nil, fmt.Errorf("failed to collect subnets: %w", err)
	}
	resources.Subnets = subnets

	// Collect Instances
	instances, err := p.collectInstances(ec2Client)
	if err != nil {
		return nil, fmt.Errorf("failed to collect instances: %w", err)
	}
	resources.Instances = instances

	// Collect Security Groups
	securityGroups, err := p.collectSecurityGroups(ec2Client)
	if err != nil {
		return nil, fmt.Errorf("failed to collect security groups: %w", err)
	}
	resources.SecurityGroups = securityGroups

	// Collect Load Balancers
	loadBalancers, err := p.collectLoadBalancers(elbClient)
	if err != nil {
		// Log but don't fail - user may not have ELB permissions
		loadBalancers = []types.LoadBalancer{}
	}
	resources.LoadBalancers = loadBalancers

	// Collect RDS Instances
	rdsInstances, err := p.collectRDSInstances(rdsClient)
	if err != nil {
		rdsInstances = []types.RDSInstance{}
	}
	resources.RDSInstances = rdsInstances

	// Collect NAT Gateways
	natGateways, err := p.collectNATGateways(ec2Client)
	if err != nil {
		natGateways = []types.NATGateway{}
	}
	resources.NATGateways = natGateways

	// Collect Internet Gateways
	internetGateways, err := p.collectInternetGateways(ec2Client)
	if err != nil {
		internetGateways = []types.InternetGateway{}
	}
	resources.InternetGateways = internetGateways

	// Collect Route Tables
	routeTables, err := p.collectRouteTables(ec2Client)
	if err != nil {
		routeTables = []types.RouteTable{}
	}
	resources.RouteTables = routeTables

	// Collect VPC Peerings
	vpcPeerings, err := p.collectVPCPeerings(ec2Client)
	if err != nil {
		vpcPeerings = []types.VPCPeering{}
	}
	resources.VPCPeerings = vpcPeerings

	// Collect Lambda Functions
	lambdaFunctions, err := p.collectLambdaFunctions(lambdaClient)
	if err != nil {
		lambdaFunctions = []types.LambdaFunction{}
	}
	resources.LambdaFunctions = lambdaFunctions

	// Collect EKS Clusters
	eksClusters, err := p.collectEKSClusters(eksClient)
	if err != nil {
		eksClusters = []types.EKSCluster{}
	}
	resources.EKSClusters = eksClusters

	// Collect ECS Clusters
	ecsClusters, err := p.collectECSClusters(ecsClient)
	if err != nil {
		ecsClusters = []types.ECSCluster{}
	}
	resources.ECSClusters = ecsClusters

	return resources, nil
}

func (p *AWSProvider) collectVPCs(client *ec2.Client) ([]types.VPC, error) {
	resp, err := client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}

	var vpcs []types.VPC
	for _, vpc := range resp.Vpcs {
		name := ""
		for _, tag := range vpc.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}

		// Check if flow logs are enabled
		flowLogsResp, _ := client.DescribeFlowLogs(context.TODO(), &ec2.DescribeFlowLogsInput{
			Filter: []ec2types.Filter{
				{
					Name:   aws.String("resource-id"),
					Values: []string{aws.ToString(vpc.VpcId)},
				},
			},
		})
		flowLogsEnabled := len(flowLogsResp.FlowLogs) > 0

		vpcs = append(vpcs, types.VPC{
			ID:              aws.ToString(vpc.VpcId),
			CIDR:            aws.ToString(vpc.CidrBlock),
			Name:            name,
			IsDefault:       aws.ToBool(vpc.IsDefault),
			FlowLogsEnabled: flowLogsEnabled,
		})
	}
	return vpcs, nil
}

func (p *AWSProvider) collectSubnets(client *ec2.Client) ([]types.Subnet, error) {
	resp, err := client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, err
	}

	// Get route tables to determine if subnet is public
	rtResp, _ := client.DescribeRouteTables(context.TODO(), &ec2.DescribeRouteTablesInput{})
	publicSubnets := make(map[string]bool)

	for _, rt := range rtResp.RouteTables {
		hasIGW := false
		for _, route := range rt.Routes {
			if route.GatewayId != nil && len(*route.GatewayId) > 0 && (*route.GatewayId)[:3] == "igw" {
				hasIGW = true
				break
			}
		}
		if hasIGW {
			for _, assoc := range rt.Associations {
				if assoc.SubnetId != nil {
					publicSubnets[*assoc.SubnetId] = true
				}
			}
		}
	}

	var subnets []types.Subnet
	for _, subnet := range resp.Subnets {
		name := ""
		for _, tag := range subnet.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}
		subnets = append(subnets, types.Subnet{
			ID:       aws.ToString(subnet.SubnetId),
			VPCID:    aws.ToString(subnet.VpcId),
			CIDR:     aws.ToString(subnet.CidrBlock),
			AZ:       aws.ToString(subnet.AvailabilityZone),
			Name:     name,
			IsPublic: publicSubnets[aws.ToString(subnet.SubnetId)],
		})
	}
	return subnets, nil
}

func (p *AWSProvider) collectInstances(client *ec2.Client) ([]types.Instance, error) {
	resp, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	var instances []types.Instance
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			name := ""
			for _, tag := range inst.Tags {
				if aws.ToString(tag.Key) == "Name" {
					name = aws.ToString(tag.Value)
					break
				}
			}

			publicIP := "N/A"
			if inst.PublicIpAddress != nil {
				publicIP = aws.ToString(inst.PublicIpAddress)
			}

			var sgIDs []string
			for _, sg := range inst.SecurityGroups {
				sgIDs = append(sgIDs, aws.ToString(sg.GroupId))
			}

			// Check IMDSv2
			imdsv2Required := false
			if inst.MetadataOptions != nil && inst.MetadataOptions.HttpTokens == ec2types.HttpTokensStateRequired {
				imdsv2Required = true
			}

			// Get IAM role if any
			iamRole := ""
			if inst.IamInstanceProfile != nil {
				iamRole = aws.ToString(inst.IamInstanceProfile.Arn)
			}

			// Check EBS encryption (simplified - checking root volume)
			ebsEncrypted := false
			for _, bdm := range inst.BlockDeviceMappings {
				if bdm.Ebs != nil {
					// Query the volume to check encryption
					volResp, err := client.DescribeVolumes(context.TODO(), &ec2.DescribeVolumesInput{
						VolumeIds: []string{aws.ToString(bdm.Ebs.VolumeId)},
					})
					if err == nil && len(volResp.Volumes) > 0 {
						ebsEncrypted = aws.ToBool(volResp.Volumes[0].Encrypted)
					}
					break // Just check the first/root volume
				}
			}

			instances = append(instances, types.Instance{
				ID:             aws.ToString(inst.InstanceId),
				VPCID:          aws.ToString(inst.VpcId),
				SubnetID:       aws.ToString(inst.SubnetId),
				PrivateIP:      aws.ToString(inst.PrivateIpAddress),
				PublicIP:       publicIP,
				SGIDs:          sgIDs,
				Name:           name,
				InstanceType:   string(inst.InstanceType),
				IMDSv2Required: imdsv2Required,
				IAMRole:        iamRole,
				EBSEncrypted:   ebsEncrypted,
			})
		}
	}
	return instances, nil
}

func (p *AWSProvider) collectSecurityGroups(client *ec2.Client) ([]types.SecurityGroup, error) {
	resp, err := client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, err
	}

	var securityGroups []types.SecurityGroup
	for _, sg := range resp.SecurityGroups {
		var ingressRules []types.SGRule
		for _, perm := range sg.IpPermissions {
			var ipRanges []string
			for _, ipRange := range perm.IpRanges {
				ipRanges = append(ipRanges, aws.ToString(ipRange.CidrIp))
			}
			ingressRules = append(ingressRules, types.SGRule{
				Direction: "ingress",
				FromPort:  aws.ToInt32(perm.FromPort),
				ToPort:    aws.ToInt32(perm.ToPort),
				Protocol:  aws.ToString(perm.IpProtocol),
				IPRanges:  ipRanges,
			})
		}

		var egressRules []types.SGRule
		for _, perm := range sg.IpPermissionsEgress {
			var ipRanges []string
			for _, ipRange := range perm.IpRanges {
				ipRanges = append(ipRanges, aws.ToString(ipRange.CidrIp))
			}
			egressRules = append(egressRules, types.SGRule{
				Direction: "egress",
				FromPort:  aws.ToInt32(perm.FromPort),
				ToPort:    aws.ToInt32(perm.ToPort),
				Protocol:  aws.ToString(perm.IpProtocol),
				IPRanges:  ipRanges,
			})
		}

		securityGroups = append(securityGroups, types.SecurityGroup{
			ID:           aws.ToString(sg.GroupId),
			VPCID:        aws.ToString(sg.VpcId),
			Name:         aws.ToString(sg.GroupName),
			Description:  aws.ToString(sg.Description),
			IngressRules: ingressRules,
			EgressRules:  egressRules,
		})
	}
	return securityGroups, nil
}

func (p *AWSProvider) collectLoadBalancers(client *elbv2.Client) ([]types.LoadBalancer, error) {
	resp, err := client.DescribeLoadBalancers(context.TODO(), &elbv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}

	var loadBalancers []types.LoadBalancer
	for _, lb := range resp.LoadBalancers {
		var subnetIDs []string
		for _, az := range lb.AvailabilityZones {
			if az.SubnetId != nil {
				subnetIDs = append(subnetIDs, *az.SubnetId)
			}
		}

		loadBalancers = append(loadBalancers, types.LoadBalancer{
			ARN:       aws.ToString(lb.LoadBalancerArn),
			Name:      aws.ToString(lb.LoadBalancerName),
			VPCID:     aws.ToString(lb.VpcId),
			Scheme:    string(lb.Scheme),
			Type:      string(lb.Type),
			SubnetIDs: subnetIDs,
		})
	}
	return loadBalancers, nil
}

func (p *AWSProvider) collectRDSInstances(client *rds.Client) ([]types.RDSInstance, error) {
	resp, err := client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, err
	}

	var rdsInstances []types.RDSInstance
	for _, db := range resp.DBInstances {
		var sgIDs []string
		for _, sg := range db.VpcSecurityGroups {
			sgIDs = append(sgIDs, aws.ToString(sg.VpcSecurityGroupId))
		}

		vpcID := ""
		subnetGroup := ""
		if db.DBSubnetGroup != nil {
			vpcID = aws.ToString(db.DBSubnetGroup.VpcId)
			subnetGroup = aws.ToString(db.DBSubnetGroup.DBSubnetGroupName)
		}

		rdsInstances = append(rdsInstances, types.RDSInstance{
			ID:                 aws.ToString(db.DBInstanceIdentifier),
			VPCID:              vpcID,
			Engine:             aws.ToString(db.Engine),
			EngineVersion:      aws.ToString(db.EngineVersion),
			InstanceClass:      aws.ToString(db.DBInstanceClass),
			PubliclyAccessible: aws.ToBool(db.PubliclyAccessible),
			Encrypted:          aws.ToBool(db.StorageEncrypted),
			SubnetGroupName:    subnetGroup,
			SecurityGroupIDs:   sgIDs,
			Name:               aws.ToString(db.DBInstanceIdentifier),
		})
	}
	return rdsInstances, nil
}

func (p *AWSProvider) collectNATGateways(client *ec2.Client) ([]types.NATGateway, error) {
	resp, err := client.DescribeNatGateways(context.TODO(), &ec2.DescribeNatGatewaysInput{})
	if err != nil {
		return nil, err
	}

	var natGateways []types.NATGateway
	for _, nat := range resp.NatGateways {
		name := ""
		for _, tag := range nat.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}

		publicIP := ""
		privateIP := ""
		for _, addr := range nat.NatGatewayAddresses {
			if addr.PublicIp != nil {
				publicIP = *addr.PublicIp
			}
			if addr.PrivateIp != nil {
				privateIP = *addr.PrivateIp
			}
		}

		natGateways = append(natGateways, types.NATGateway{
			ID:        aws.ToString(nat.NatGatewayId),
			VPCID:     aws.ToString(nat.VpcId),
			SubnetID:  aws.ToString(nat.SubnetId),
			PublicIP:  publicIP,
			PrivateIP: privateIP,
			State:     string(nat.State),
			Name:      name,
		})
	}
	return natGateways, nil
}

func (p *AWSProvider) collectInternetGateways(client *ec2.Client) ([]types.InternetGateway, error) {
	resp, err := client.DescribeInternetGateways(context.TODO(), &ec2.DescribeInternetGatewaysInput{})
	if err != nil {
		return nil, err
	}

	var igws []types.InternetGateway
	for _, igw := range resp.InternetGateways {
		name := ""
		for _, tag := range igw.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}

		var vpcIDs []string
		for _, attach := range igw.Attachments {
			vpcIDs = append(vpcIDs, aws.ToString(attach.VpcId))
		}

		igws = append(igws, types.InternetGateway{
			ID:     aws.ToString(igw.InternetGatewayId),
			VPCIDs: vpcIDs,
			Name:   name,
		})
	}
	return igws, nil
}

func (p *AWSProvider) collectRouteTables(client *ec2.Client) ([]types.RouteTable, error) {
	resp, err := client.DescribeRouteTables(context.TODO(), &ec2.DescribeRouteTablesInput{})
	if err != nil {
		return nil, err
	}

	var routeTables []types.RouteTable
	for _, rt := range resp.RouteTables {
		name := ""
		for _, tag := range rt.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}

		isMain := false
		for _, assoc := range rt.Associations {
			if aws.ToBool(assoc.Main) {
				isMain = true
				break
			}
		}

		var routes []types.Route
		for _, route := range rt.Routes {
			routes = append(routes, types.Route{
				DestinationCIDR: aws.ToString(route.DestinationCidrBlock),
				GatewayID:       aws.ToString(route.GatewayId),
				NATGatewayID:    aws.ToString(route.NatGatewayId),
				InstanceID:      aws.ToString(route.InstanceId),
				State:           string(route.State),
			})
		}

		routeTables = append(routeTables, types.RouteTable{
			ID:     aws.ToString(rt.RouteTableId),
			VPCID:  aws.ToString(rt.VpcId),
			Name:   name,
			Routes: routes,
			IsMain: isMain,
		})
	}
	return routeTables, nil
}

func (p *AWSProvider) collectVPCPeerings(client *ec2.Client) ([]types.VPCPeering, error) {
	resp, err := client.DescribeVpcPeeringConnections(context.TODO(), &ec2.DescribeVpcPeeringConnectionsInput{})
	if err != nil {
		return nil, err
	}

	var peerings []types.VPCPeering
	for _, pc := range resp.VpcPeeringConnections {
		name := ""
		for _, tag := range pc.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}

		requesterVPC := ""
		accepterVPC := ""
		if pc.RequesterVpcInfo != nil {
			requesterVPC = aws.ToString(pc.RequesterVpcInfo.VpcId)
		}
		if pc.AccepterVpcInfo != nil {
			accepterVPC = aws.ToString(pc.AccepterVpcInfo.VpcId)
		}

		status := ""
		if pc.Status != nil {
			status = string(pc.Status.Code)
		}

		peerings = append(peerings, types.VPCPeering{
			ID:             aws.ToString(pc.VpcPeeringConnectionId),
			RequesterVPCID: requesterVPC,
			AccepterVPCID:  accepterVPC,
			Status:         status,
			Name:           name,
		})
	}
	return peerings, nil
}

func (p *AWSProvider) collectLambdaFunctions(client *lambda.Client) ([]types.LambdaFunction, error) {
	resp, err := client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})
	if err != nil {
		return nil, err
	}

	var functions []types.LambdaFunction
	for _, fn := range resp.Functions {
		lf := types.LambdaFunction{
			Name:    aws.ToString(fn.FunctionName),
			ARN:     aws.ToString(fn.FunctionArn),
			Runtime: string(fn.Runtime),
		}

		if fn.VpcConfig != nil {
			lf.VPCID = aws.ToString(fn.VpcConfig.VpcId)
			lf.SubnetIDs = fn.VpcConfig.SubnetIds
			lf.SecurityGroupIDs = fn.VpcConfig.SecurityGroupIds
		}

		functions = append(functions, lf)
	}
	return functions, nil
}

func (p *AWSProvider) collectEKSClusters(client *eks.Client) ([]types.EKSCluster, error) {
	listResp, err := client.ListClusters(context.TODO(), &eks.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	var clusters []types.EKSCluster
	for _, clusterName := range listResp.Clusters {
		descResp, err := client.DescribeCluster(context.TODO(), &eks.DescribeClusterInput{
			Name: aws.String(clusterName),
		})
		if err != nil {
			continue
		}

		cluster := descResp.Cluster
		ec := types.EKSCluster{
			Name:   aws.ToString(cluster.Name),
			ARN:    aws.ToString(cluster.Arn),
			Status: string(cluster.Status),
		}

		if cluster.ResourcesVpcConfig != nil {
			ec.VPCID = aws.ToString(cluster.ResourcesVpcConfig.VpcId)
			ec.SubnetIDs = cluster.ResourcesVpcConfig.SubnetIds
			ec.SecurityGroupID = aws.ToString(cluster.ResourcesVpcConfig.ClusterSecurityGroupId)
		}

		clusters = append(clusters, ec)
	}
	return clusters, nil
}

func (p *AWSProvider) collectECSClusters(client *ecs.Client) ([]types.ECSCluster, error) {
	listResp, err := client.ListClusters(context.TODO(), &ecs.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	if len(listResp.ClusterArns) == 0 {
		return []types.ECSCluster{}, nil
	}

	descResp, err := client.DescribeClusters(context.TODO(), &ecs.DescribeClustersInput{
		Clusters: listResp.ClusterArns,
	})
	if err != nil {
		return nil, err
	}

	var clusters []types.ECSCluster
	for _, cluster := range descResp.Clusters {
		clusters = append(clusters, types.ECSCluster{
			Name:         aws.ToString(cluster.ClusterName),
			ARN:          aws.ToString(cluster.ClusterArn),
			Status:       aws.ToString(cluster.Status),
			ServiceCount: int(cluster.ActiveServicesCount),
			TaskCount:    int(cluster.RunningTasksCount),
		})
	}
	return clusters, nil
}
