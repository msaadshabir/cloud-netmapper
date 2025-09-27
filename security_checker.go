package main

import "fmt"

type Risk struct {
	Type       string
	Resource   string
	Details    string
	Severity   string
}

func checkSecurityRisks(resources *AWSResources) []Risk {
	var risks []Risk

	for _, sg := range resources.SecurityGroups {
		for _, rule := range sg.Rules {
			for _, cidr := range rule.IPRanges {
				if cidr == "0.0.0.0/0" {
					port := rule.FromPort
					if port == 22 || port == 3389 || port == 21 || port == 23 || port == 0 {
						risks = append(risks, Risk{
							Type:     "Open Security Group",
							Resource: sg.Name,
							Details:  fmt.Sprintf("Port %d open to 0.0.0.0/0", port),
							Severity: "High",
						})
					}
				}
			}
		}
	}

	hasPublicInstances := false
	for _, inst := range resources.Instances {
		if inst.PublicIP != "N/A" {
			hasPublicInstances = true
			break
		}
	}
	if hasPublicInstances && len(resources.LoadBalancers) == 0 {
		risks = append(risks, Risk{
			Type:     "Direct Public Exposure",
			Resource: "EC2 Instances",
			Details:  "Instances exposed directly to internet (no load balancer)",
			Severity: "Medium",
		})
	}

	return risks
}
