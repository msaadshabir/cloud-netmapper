package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type VPC struct {
	ID   string
	CIDR string
	Name string
}

type Subnet struct {
	ID    string
	VPCID string
	CIDR  string
	AZ    string
	Name  string
}

type Instance struct {
	ID        string
	VPCID     string
	SubnetID  string
	PrivateIP string
	PublicIP  string
	SGIDs     []string
	Name      string
}

type SecurityGroup struct {
	ID          string
	Name        string
	Description string
	Rules       []SGRule
}

type SGRule struct {
	FromPort int32
	IPRanges []string
}

type LoadBalancer struct {
	ARN    string
	Name   string
	VPCID  string
	Scheme string
	Type   string
}

type AWSResources struct {
	VPCs           []VPC
	Subnets        []Subnet
	Instances      []Instance
	SecurityGroups []SecurityGroup
	LoadBalancers  []LoadBalancer
}

func getAWSResources(region string) (*AWSResources, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	resources := &AWSResources{}

	vpcResp, err := ec2Client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %v", err)
	}
	for _, vpc := range vpcResp.Vpcs {
		if vpc.VpcId != nil && vpc.CidrBlock != nil {
			resources.VPCs = append(resources.VPCs, VPC{
				ID:   *vpc.VpcId,
				CIDR: *vpc.CidrBlock,
				Name: getNameTag(vpc.Tags),
			})
		}
	}

	subnetResp, err := ec2Client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %v", err)
	}
	for _, subnet := range subnetResp.Subnets {
		if subnet.SubnetId != nil && subnet.VpcId != nil &&
			subnet.CidrBlock != nil && subnet.AvailabilityZone != nil {
			resources.Subnets = append(resources.Subnets, Subnet{
				ID:    *subnet.SubnetId,
				VPCID: *subnet.VpcId,
				CIDR:  *subnet.CidrBlock,
				AZ:    *subnet.AvailabilityZone,
				Name:  getNameTag(subnet.Tags),
			})
		}
	}

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

				// Skip instances with missing required fields
				if instance.InstanceId == nil || instance.VpcId == nil ||
					instance.SubnetId == nil || instance.PrivateIpAddress == nil {
					continue
				}

				resources.Instances = append(resources.Instances, Instance{
					ID:        *instance.InstanceId,
					VPCID:     *instance.VpcId,
					SubnetID:  *instance.SubnetId,
					PrivateIP: *instance.PrivateIpAddress,
					PublicIP:  publicIP,
					SGIDs:     sgIDs,
					Name:      getNameTag(instance.Tags),
				})
			}
		}
	}

	sgResp, err := ec2Client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe security groups: %v", err)
	}
	for _, sg := range sgResp.SecurityGroups {
		if sg.GroupId == nil || sg.GroupName == nil {
			continue
		}

		rules := []SGRule{}
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
			rules = append(rules, SGRule{
				FromPort: fromPort,
				IPRanges: ipRanges,
			})
		}

		description := ""
		if sg.Description != nil {
			description = *sg.Description
		}

		resources.SecurityGroups = append(resources.SecurityGroups, SecurityGroup{
			ID:          *sg.GroupId,
			Name:        *sg.GroupName,
			Description: description,
			Rules:       rules,
		})
	}

	lbResp, err := elbClient.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		log.Printf("Warning: no load balancers found or access denied: %v", err)
	} else {
		for _, lb := range lbResp.LoadBalancers {
			if lb.LoadBalancerArn != nil && lb.LoadBalancerName != nil && lb.VpcId != nil {
				resources.LoadBalancers = append(resources.LoadBalancers, LoadBalancer{
					ARN:    *lb.LoadBalancerArn,
					Name:   *lb.LoadBalancerName,
					VPCID:  *lb.VpcId,
					Scheme: string(lb.Scheme),
					Type:   string(lb.Type),
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
