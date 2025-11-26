package main

import (
	"context"
	"fmt"

	"cloud-netmapper/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

type VPC struct {
	ID              string
	CIDR            string
	Name            string
	IsDefault       bool
	FlowLogsEnabled bool
}

type Subnet struct {
	ID       string
	VPCID    string
	CIDR     string
	AZ       string
	Name     string
	IsPublic bool
}

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

type SecurityGroup struct {
	ID           string
	VPCID        string
	Name         string
	Description  string
	IngressRules []SGRule
	EgressRules  []SGRule
}

type SGRule struct {
	Direction   string
	FromPort    int32
	ToPort      int32
	Protocol    string
	IPRanges    []string
	Description string
}

type LoadBalancer struct {
	ARN       string
	Name      string
	VPCID     string
	Scheme    string
	Type      string
	SubnetIDs []string
}

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

type NATGateway struct {
	ID        string
	VPCID     string
	SubnetID  string
	PublicIP  string
	PrivateIP string
	State     string
	Name      string
}

type InternetGateway struct {
	ID     string
	VPCIDs []string
	Name   string
}

type RouteTable struct {
	ID     string
	VPCID  string
	Name   string
	Routes []Route
	IsMain bool
}

type Route struct {
	DestinationCIDR string
	GatewayID       string
	NATGatewayID    string
	InstanceID      string
	State           string
}

type VPCPeering struct {
	ID             string
	RequesterVPCID string
	AccepterVPCID  string
	Status         string
	Name           string
}

type LambdaFunction struct {
	Name             string
	ARN              string
	Runtime          string
	VPCID            string
	SubnetIDs        []string
	SecurityGroupIDs []string
}

type EKSCluster struct {
	Name            string
	ARN             string
	VPCID           string
	SubnetIDs       []string
	SecurityGroupID string
	Status          string
}

type ECSCluster struct {
	Name         string
	ARN          string
	Status       string
	ServiceCount int
	TaskCount    int
}

type AWSResources struct {
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

func getAWSResources(region string) (*AWSResources, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)
	rdsClient := rds.NewFromConfig(cfg)
	lambdaClient := lambda.NewFromConfig(cfg)
	eksClient := eks.NewFromConfig(cfg)
	ecsClient := ecs.NewFromConfig(cfg)

	resources := &AWSResources{}

	// Collect VPCs with flow logs check
	vpcResp, err := ec2Client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %v", err)
	}

	// Get flow logs to check which VPCs have them enabled
	flowLogsResp, _ := ec2Client.DescribeFlowLogs(context.TODO(), &ec2.DescribeFlowLogsInput{})
	flowLogVPCs := make(map[string]bool)
	if flowLogsResp != nil {
		for _, fl := range flowLogsResp.FlowLogs {
			if fl.ResourceId != nil {
				flowLogVPCs[*fl.ResourceId] = true
			}
		}
	}

	for _, vpc := range vpcResp.Vpcs {
		if vpc.VpcId != nil && vpc.CidrBlock != nil {
			isDefault := false
			if vpc.IsDefault != nil {
				isDefault = *vpc.IsDefault
			}
			resources.VPCs = append(resources.VPCs, VPC{
				ID:              *vpc.VpcId,
				CIDR:            *vpc.CidrBlock,
				Name:            getNameTag(vpc.Tags),
				IsDefault:       isDefault,
				FlowLogsEnabled: flowLogVPCs[*vpc.VpcId],
			})
		}
	}

	// Collect Subnets
	subnetResp, err := ec2Client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %v", err)
	}
	for _, subnet := range subnetResp.Subnets {
		if subnet.SubnetId != nil && subnet.VpcId != nil &&
			subnet.CidrBlock != nil && subnet.AvailabilityZone != nil {
			isPublic := false
			if subnet.MapPublicIpOnLaunch != nil {
				isPublic = *subnet.MapPublicIpOnLaunch
			}
			resources.Subnets = append(resources.Subnets, Subnet{
				ID:       *subnet.SubnetId,
				VPCID:    *subnet.VpcId,
				CIDR:     *subnet.CidrBlock,
				AZ:       *subnet.AvailabilityZone,
				Name:     getNameTag(subnet.Tags),
				IsPublic: isPublic,
			})
		}
	}

	// Collect EC2 Instances with additional metadata
	instanceResp, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %v", err)
	}
	for _, reservation := range instanceResp.Reservations {
		for _, instance := range reservation.Instances {
			if instance.State != nil && instance.State.Name == "running" {
				sgIDs := []string{}
				for _, sg := range instance.SecurityGroups {
					if sg.GroupId != nil {
						sgIDs = append(sgIDs, *sg.GroupId)
					}
				}
				publicIP := "N/A"
				if instance.PublicIpAddress != nil {
					publicIP = *instance.PublicIpAddress
				}

				if instance.InstanceId == nil || instance.VpcId == nil ||
					instance.SubnetId == nil || instance.PrivateIpAddress == nil {
					continue
				}

				instanceType := ""
				if instance.InstanceType != "" {
					instanceType = string(instance.InstanceType)
				}

				imdsv2Required := false
				if instance.MetadataOptions != nil && instance.MetadataOptions.HttpTokens == "required" {
					imdsv2Required = true
				}

				iamRole := ""
				if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
					iamRole = *instance.IamInstanceProfile.Arn
				}

				ebsEncrypted := true
				for _, bdm := range instance.BlockDeviceMappings {
					if bdm.Ebs != nil {
						// Check if EBS volume is encrypted
						if bdm.Ebs.VolumeId != nil {
							volResp, err := ec2Client.DescribeVolumes(context.TODO(), &ec2.DescribeVolumesInput{
								VolumeIds: []string{*bdm.Ebs.VolumeId},
							})
							if err == nil && len(volResp.Volumes) > 0 {
								if volResp.Volumes[0].Encrypted != nil && !*volResp.Volumes[0].Encrypted {
									ebsEncrypted = false
								}
							}
						}
					}
				}

				resources.Instances = append(resources.Instances, Instance{
					ID:             *instance.InstanceId,
					VPCID:          *instance.VpcId,
					SubnetID:       *instance.SubnetId,
					PrivateIP:      *instance.PrivateIpAddress,
					PublicIP:       publicIP,
					SGIDs:          sgIDs,
					Name:           getNameTag(instance.Tags),
					InstanceType:   instanceType,
					IMDSv2Required: imdsv2Required,
					IAMRole:        iamRole,
					EBSEncrypted:   ebsEncrypted,
				})
			}
		}
	}

	// Collect Security Groups with ingress and egress rules
	sgResp, err := ec2Client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe security groups: %v", err)
	}
	for _, sg := range sgResp.SecurityGroups {
		if sg.GroupId == nil || sg.GroupName == nil {
			continue
		}

		ingressRules := []SGRule{}
		for _, perm := range sg.IpPermissions {
			ipRanges := []string{}
			for _, ipr := range perm.IpRanges {
				if ipr.CidrIp != nil {
					ipRanges = append(ipRanges, *ipr.CidrIp)
				}
			}
			fromPort := int32(0)
			if perm.FromPort != nil {
				fromPort = *perm.FromPort
			}
			toPort := int32(0)
			if perm.ToPort != nil {
				toPort = *perm.ToPort
			}
			protocol := "all"
			if perm.IpProtocol != nil {
				protocol = *perm.IpProtocol
			}
			ingressRules = append(ingressRules, SGRule{
				Direction: "ingress",
				FromPort:  fromPort,
				ToPort:    toPort,
				Protocol:  protocol,
				IPRanges:  ipRanges,
			})
		}

		egressRules := []SGRule{}
		for _, perm := range sg.IpPermissionsEgress {
			ipRanges := []string{}
			for _, ipr := range perm.IpRanges {
				if ipr.CidrIp != nil {
					ipRanges = append(ipRanges, *ipr.CidrIp)
				}
			}
			fromPort := int32(0)
			if perm.FromPort != nil {
				fromPort = *perm.FromPort
			}
			toPort := int32(0)
			if perm.ToPort != nil {
				toPort = *perm.ToPort
			}
			protocol := "all"
			if perm.IpProtocol != nil {
				protocol = *perm.IpProtocol
			}
			egressRules = append(egressRules, SGRule{
				Direction: "egress",
				FromPort:  fromPort,
				ToPort:    toPort,
				Protocol:  protocol,
				IPRanges:  ipRanges,
			})
		}

		description := ""
		if sg.Description != nil {
			description = *sg.Description
		}

		vpcID := ""
		if sg.VpcId != nil {
			vpcID = *sg.VpcId
		}

		resources.SecurityGroups = append(resources.SecurityGroups, SecurityGroup{
			ID:           *sg.GroupId,
			VPCID:        vpcID,
			Name:         *sg.GroupName,
			Description:  description,
			IngressRules: ingressRules,
			EgressRules:  egressRules,
		})
	}

	// Collect Load Balancers
	lbResp, err := elbClient.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		logger.Warn("Failed to describe load balancers", "error", err)
	} else {
		for _, lb := range lbResp.LoadBalancers {
			if lb.LoadBalancerArn != nil && lb.LoadBalancerName != nil && lb.VpcId != nil {
				subnetIDs := []string{}
				for _, az := range lb.AvailabilityZones {
					if az.SubnetId != nil {
						subnetIDs = append(subnetIDs, *az.SubnetId)
					}
				}
				resources.LoadBalancers = append(resources.LoadBalancers, LoadBalancer{
					ARN:       *lb.LoadBalancerArn,
					Name:      *lb.LoadBalancerName,
					VPCID:     *lb.VpcId,
					Scheme:    string(lb.Scheme),
					Type:      string(lb.Type),
					SubnetIDs: subnetIDs,
				})
			}
		}
	}

	// Collect NAT Gateways
	natResp, err := ec2Client.DescribeNatGateways(context.TODO(), &ec2.DescribeNatGatewaysInput{})
	if err != nil {
		logger.Warn("Failed to describe NAT gateways", "error", err)
	} else {
		for _, nat := range natResp.NatGateways {
			if nat.NatGatewayId == nil {
				continue
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
			vpcID := ""
			if nat.VpcId != nil {
				vpcID = *nat.VpcId
			}
			subnetID := ""
			if nat.SubnetId != nil {
				subnetID = *nat.SubnetId
			}
			resources.NATGateways = append(resources.NATGateways, NATGateway{
				ID:        *nat.NatGatewayId,
				VPCID:     vpcID,
				SubnetID:  subnetID,
				PublicIP:  publicIP,
				PrivateIP: privateIP,
				State:     string(nat.State),
				Name:      getNameTag(nat.Tags),
			})
		}
	}

	// Collect Internet Gateways
	igwResp, err := ec2Client.DescribeInternetGateways(context.TODO(), &ec2.DescribeInternetGatewaysInput{})
	if err != nil {
		logger.Warn("Failed to describe internet gateways", "error", err)
	} else {
		for _, igw := range igwResp.InternetGateways {
			if igw.InternetGatewayId == nil {
				continue
			}
			vpcIDs := []string{}
			for _, att := range igw.Attachments {
				if att.VpcId != nil {
					vpcIDs = append(vpcIDs, *att.VpcId)
				}
			}
			resources.InternetGateways = append(resources.InternetGateways, InternetGateway{
				ID:     *igw.InternetGatewayId,
				VPCIDs: vpcIDs,
				Name:   getNameTag(igw.Tags),
			})
		}
	}

	// Collect Route Tables
	rtResp, err := ec2Client.DescribeRouteTables(context.TODO(), &ec2.DescribeRouteTablesInput{})
	if err != nil {
		logger.Warn("Failed to describe route tables", "error", err)
	} else {
		for _, rt := range rtResp.RouteTables {
			if rt.RouteTableId == nil {
				continue
			}
			vpcID := ""
			if rt.VpcId != nil {
				vpcID = *rt.VpcId
			}
			routes := []Route{}
			for _, r := range rt.Routes {
				destCIDR := ""
				if r.DestinationCidrBlock != nil {
					destCIDR = *r.DestinationCidrBlock
				}
				gwID := ""
				if r.GatewayId != nil {
					gwID = *r.GatewayId
				}
				natGwID := ""
				if r.NatGatewayId != nil {
					natGwID = *r.NatGatewayId
				}
				instID := ""
				if r.InstanceId != nil {
					instID = *r.InstanceId
				}
				routes = append(routes, Route{
					DestinationCIDR: destCIDR,
					GatewayID:       gwID,
					NATGatewayID:    natGwID,
					InstanceID:      instID,
					State:           string(r.State),
				})
			}
			isMain := false
			for _, assoc := range rt.Associations {
				if assoc.Main != nil && *assoc.Main {
					isMain = true
					break
				}
			}
			resources.RouteTables = append(resources.RouteTables, RouteTable{
				ID:     *rt.RouteTableId,
				VPCID:  vpcID,
				Name:   getNameTag(rt.Tags),
				Routes: routes,
				IsMain: isMain,
			})
		}
	}

	// Collect VPC Peering Connections
	peerResp, err := ec2Client.DescribeVpcPeeringConnections(context.TODO(), &ec2.DescribeVpcPeeringConnectionsInput{})
	if err != nil {
		logger.Warn("Failed to describe VPC peering connections", "error", err)
	} else {
		for _, peer := range peerResp.VpcPeeringConnections {
			if peer.VpcPeeringConnectionId == nil {
				continue
			}
			reqVPC := ""
			accVPC := ""
			if peer.RequesterVpcInfo != nil && peer.RequesterVpcInfo.VpcId != nil {
				reqVPC = *peer.RequesterVpcInfo.VpcId
			}
			if peer.AccepterVpcInfo != nil && peer.AccepterVpcInfo.VpcId != nil {
				accVPC = *peer.AccepterVpcInfo.VpcId
			}
			status := ""
			if peer.Status != nil && peer.Status.Code != "" {
				status = string(peer.Status.Code)
			}
			resources.VPCPeerings = append(resources.VPCPeerings, VPCPeering{
				ID:             *peer.VpcPeeringConnectionId,
				RequesterVPCID: reqVPC,
				AccepterVPCID:  accVPC,
				Status:         status,
				Name:           getNameTag(peer.Tags),
			})
		}
	}

	// Collect RDS Instances
	rdsResp, err := rdsClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		logger.Warn("Failed to describe RDS instances", "error", err)
	} else {
		for _, db := range rdsResp.DBInstances {
			if db.DBInstanceIdentifier == nil {
				continue
			}
			vpcID := ""
			if db.DBSubnetGroup != nil && db.DBSubnetGroup.VpcId != nil {
				vpcID = *db.DBSubnetGroup.VpcId
			}
			sgIDs := []string{}
			for _, sg := range db.VpcSecurityGroups {
				if sg.VpcSecurityGroupId != nil {
					sgIDs = append(sgIDs, *sg.VpcSecurityGroupId)
				}
			}
			engine := ""
			if db.Engine != nil {
				engine = *db.Engine
			}
			engineVersion := ""
			if db.EngineVersion != nil {
				engineVersion = *db.EngineVersion
			}
			instanceClass := ""
			if db.DBInstanceClass != nil {
				instanceClass = *db.DBInstanceClass
			}
			encrypted := false
			if db.StorageEncrypted != nil {
				encrypted = *db.StorageEncrypted
			}
			publiclyAccessible := false
			if db.PubliclyAccessible != nil {
				publiclyAccessible = *db.PubliclyAccessible
			}
			subnetGroupName := ""
			if db.DBSubnetGroup != nil && db.DBSubnetGroup.DBSubnetGroupName != nil {
				subnetGroupName = *db.DBSubnetGroup.DBSubnetGroupName
			}
			resources.RDSInstances = append(resources.RDSInstances, RDSInstance{
				ID:                 *db.DBInstanceIdentifier,
				VPCID:              vpcID,
				Engine:             engine,
				EngineVersion:      engineVersion,
				InstanceClass:      instanceClass,
				PubliclyAccessible: publiclyAccessible,
				Encrypted:          encrypted,
				SubnetGroupName:    subnetGroupName,
				SecurityGroupIDs:   sgIDs,
				Name:               *db.DBInstanceIdentifier,
			})
		}
	}

	// Collect Lambda Functions with VPC config
	lambdaResp, err := lambdaClient.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})
	if err != nil {
		logger.Warn("Failed to list Lambda functions", "error", err)
	} else {
		for _, fn := range lambdaResp.Functions {
			if fn.FunctionName == nil {
				continue
			}
			vpcID := ""
			subnetIDs := []string{}
			sgIDs := []string{}
			if fn.VpcConfig != nil {
				if fn.VpcConfig.VpcId != nil {
					vpcID = *fn.VpcConfig.VpcId
				}
				subnetIDs = fn.VpcConfig.SubnetIds
				sgIDs = fn.VpcConfig.SecurityGroupIds
			}
			runtime := ""
			if fn.Runtime != "" {
				runtime = string(fn.Runtime)
			}
			arn := ""
			if fn.FunctionArn != nil {
				arn = *fn.FunctionArn
			}
			resources.LambdaFunctions = append(resources.LambdaFunctions, LambdaFunction{
				Name:             *fn.FunctionName,
				ARN:              arn,
				Runtime:          runtime,
				VPCID:            vpcID,
				SubnetIDs:        subnetIDs,
				SecurityGroupIDs: sgIDs,
			})
		}
	}

	// Collect EKS Clusters
	eksListResp, err := eksClient.ListClusters(context.TODO(), &eks.ListClustersInput{})
	if err != nil {
		logger.Warn("Failed to list EKS clusters", "error", err)
	} else {
		for _, clusterName := range eksListResp.Clusters {
			eksDescResp, err := eksClient.DescribeCluster(context.TODO(), &eks.DescribeClusterInput{
				Name: aws.String(clusterName),
			})
			if err != nil {
				logger.Warn("Failed to describe EKS cluster", "cluster", clusterName, "error", err)
				continue
			}
			if eksDescResp.Cluster == nil {
				continue
			}
			cluster := eksDescResp.Cluster
			vpcID := ""
			subnetIDs := []string{}
			sgID := ""
			if cluster.ResourcesVpcConfig != nil {
				if cluster.ResourcesVpcConfig.VpcId != nil {
					vpcID = *cluster.ResourcesVpcConfig.VpcId
				}
				subnetIDs = cluster.ResourcesVpcConfig.SubnetIds
				if cluster.ResourcesVpcConfig.ClusterSecurityGroupId != nil {
					sgID = *cluster.ResourcesVpcConfig.ClusterSecurityGroupId
				}
			}
			arn := ""
			if cluster.Arn != nil {
				arn = *cluster.Arn
			}
			status := ""
			if cluster.Status != "" {
				status = string(cluster.Status)
			}
			resources.EKSClusters = append(resources.EKSClusters, EKSCluster{
				Name:            clusterName,
				ARN:             arn,
				VPCID:           vpcID,
				SubnetIDs:       subnetIDs,
				SecurityGroupID: sgID,
				Status:          status,
			})
		}
	}

	// Collect ECS Clusters
	ecsListResp, err := ecsClient.ListClusters(context.TODO(), &ecs.ListClustersInput{})
	if err != nil {
		logger.Warn("Failed to list ECS clusters", "error", err)
	} else if len(ecsListResp.ClusterArns) > 0 {
		ecsDescResp, err := ecsClient.DescribeClusters(context.TODO(), &ecs.DescribeClustersInput{
			Clusters: ecsListResp.ClusterArns,
		})
		if err != nil {
			logger.Warn("Failed to describe ECS clusters", "error", err)
		} else {
			for _, cluster := range ecsDescResp.Clusters {
				name := ""
				if cluster.ClusterName != nil {
					name = *cluster.ClusterName
				}
				arn := ""
				if cluster.ClusterArn != nil {
					arn = *cluster.ClusterArn
				}
				status := ""
				if cluster.Status != nil {
					status = *cluster.Status
				}
				resources.ECSClusters = append(resources.ECSClusters, ECSCluster{
					Name:         name,
					ARN:          arn,
					Status:       status,
					ServiceCount: int(cluster.ActiveServicesCount),
					TaskCount:    int(cluster.RunningTasksCount),
				})
			}
		}
	}

	return resources, nil
}

func getNameTag(tags []types.Tag) string {
	for _, tag := range tags {
		if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
			return *tag.Value
		}
	}
	return "Unnamed"
}
