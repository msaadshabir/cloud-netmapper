package main

import (
	"fmt"
	"os"
	"strings"
)

func generateDOTFile(resources *AWSResources, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "digraph AWS_Network {")
	fmt.Fprintln(file, "  rankdir=LR;")
	fmt.Fprintln(file, "  node [fontsize=10];")

	for _, vpc := range resources.VPCs {
		label := fmt.Sprintf("VPC\\n%s\\n%s", vpc.Name, vpc.CIDR)
		id := "vpc_" + strings.ReplaceAll(vpc.ID, ":", "_")
		fmt.Fprintf(file, "  %s [label=\"%s\", shape=box, style=filled, fillcolor=\"#E0E0E0\"];\n", id, label)
	}

	for _, subnet := range resources.Subnets {
		label := fmt.Sprintf("Subnet\\n%s\\n%s\\nAZ: %s", subnet.Name, subnet.CIDR, subnet.AZ)
		subnetID := "subnet_" + strings.ReplaceAll(subnet.ID, ":", "_")
		vpcID := "vpc_" + strings.ReplaceAll(subnet.VPCID, ":", "_")
		fmt.Fprintf(file, "  %s [label=\"%s\", shape=ellipse, style=filled, fillcolor=\"#FFD700\"];\n", subnetID, label)
		fmt.Fprintf(file, "  %s -> %s;\n", vpcID, subnetID)
	}

	for _, inst := range resources.Instances {
		ip := inst.PublicIP
		if ip == "N/A" {
			ip = inst.PrivateIP
		}
		label := fmt.Sprintf("EC2\\n%s\\n%s", inst.Name, ip)
		instID := "instance_" + strings.ReplaceAll(inst.ID, ":", "_")
		subnetID := "subnet_" + strings.ReplaceAll(inst.SubnetID, ":", "_")
		fmt.Fprintf(file, "  %s [label=\"%s\", shape=circle, style=filled, fillcolor=\"#90EE90\"];\n", instID, label)
		fmt.Fprintf(file, "  %s -> %s;\n", subnetID, instID)
	}

	for _, lb := range resources.LoadBalancers {
		label := fmt.Sprintf("LB\\n%s\\n(%s)", lb.Name, lb.Scheme)
		lbID := "lb_" + strings.ReplaceAll(lb.ARN, ":", "_")
		vpcID := "vpc_" + strings.ReplaceAll(lb.VPCID, ":", "_")
		fmt.Fprintf(file, "  %s [label=\"%s\", shape=diamond, style=filled, fillcolor=\"#FFB6C1\"];\n", lbID, label)
		fmt.Fprintf(file, "  %s -> %s;\n", vpcID, lbID)
	}

	fmt.Fprintln(file, "}")
	return nil
}
