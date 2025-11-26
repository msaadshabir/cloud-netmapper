package main

import (
	"fmt"
	"strings"
)

type Risk struct {
	Type        string
	Resource    string
	ResourceID  string
	Details     string
	Severity    string
	Remediation string
}

func checkSecurityRisks(resources *AWSResources) []Risk {
	var risks []Risk

	// Check Security Groups for open ports
	for _, sg := range resources.SecurityGroups {
		// Check ingress rules
		for _, rule := range sg.IngressRules {
			for _, cidr := range rule.IPRanges {
				if cidr == "0.0.0.0/0" {
					port := rule.FromPort
					if port == 22 || port == 3389 || port == 21 || port == 23 || port == 0 {
						portName := getPortName(port)
						risks = append(risks, Risk{
							Type:        "Open Security Group",
							Resource:    sg.Name,
							ResourceID:  sg.ID,
							Details:     fmt.Sprintf("Port %d (%s) open to 0.0.0.0/0", port, portName),
							Severity:    "high",
							Remediation: fmt.Sprintf("Restrict port %d access to specific IP ranges", port),
						})
					}
				}
			}
		}

		// Check egress rules for overly permissive access
		for _, rule := range sg.EgressRules {
			for _, cidr := range rule.IPRanges {
				if cidr == "0.0.0.0/0" && rule.Protocol == "-1" {
					risks = append(risks, Risk{
						Type:        "Overly Permissive Egress",
						Resource:    sg.Name,
						ResourceID:  sg.ID,
						Details:     "Security group allows all outbound traffic to any destination",
						Severity:    "low",
						Remediation: "Consider restricting egress rules to only required destinations",
					})
					break
				}
			}
		}
	}

	// Check for public instances without load balancers
	hasPublicInstances := false
	for _, inst := range resources.Instances {
		if inst.PublicIP != "N/A" {
			hasPublicInstances = true
			break
		}
	}
	if hasPublicInstances && len(resources.LoadBalancers) == 0 {
		risks = append(risks, Risk{
			Type:        "Direct Public Exposure",
			Resource:    "EC2 Instances",
			Details:     "Instances exposed directly to internet (no load balancer)",
			Severity:    "medium",
			Remediation: "Consider using a load balancer or NAT gateway for internet access",
		})
	}

	// Check for default VPCs in use
	for _, vpc := range resources.VPCs {
		if vpc.IsDefault {
			risks = append(risks, Risk{
				Type:        "Default VPC In Use",
				Resource:    vpc.Name,
				ResourceID:  vpc.ID,
				Details:     "Default VPC is being used which may have overly permissive settings",
				Severity:    "low",
				Remediation: "Consider creating custom VPCs with proper network segmentation",
			})
		}
	}

	// Check for VPCs without flow logs
	for _, vpc := range resources.VPCs {
		if !vpc.FlowLogsEnabled {
			risks = append(risks, Risk{
				Type:        "Missing VPC Flow Logs",
				Resource:    vpc.Name,
				ResourceID:  vpc.ID,
				Details:     "VPC does not have flow logs enabled for network monitoring",
				Severity:    "medium",
				Remediation: "Enable VPC flow logs for network traffic analysis and security monitoring",
			})
		}
	}

	// Check for instances without IMDSv2
	for _, inst := range resources.Instances {
		if !inst.IMDSv2Required {
			risks = append(risks, Risk{
				Type:        "IMDSv2 Not Enforced",
				Resource:    inst.Name,
				ResourceID:  inst.ID,
				Details:     "Instance does not require IMDSv2, vulnerable to SSRF attacks",
				Severity:    "medium",
				Remediation: "Configure instance to require IMDSv2 (HttpTokens=required)",
			})
		}
	}

	// Check for unencrypted EBS volumes
	for _, inst := range resources.Instances {
		if !inst.EBSEncrypted {
			risks = append(risks, Risk{
				Type:        "Unencrypted EBS Volume",
				Resource:    inst.Name,
				ResourceID:  inst.ID,
				Details:     "Instance has unencrypted EBS volumes attached",
				Severity:    "medium",
				Remediation: "Enable encryption for EBS volumes to protect data at rest",
			})
		}
	}

	// Check for publicly accessible RDS instances
	for _, rds := range resources.RDSInstances {
		if rds.PubliclyAccessible {
			risks = append(risks, Risk{
				Type:        "Publicly Accessible RDS",
				Resource:    rds.Name,
				ResourceID:  rds.ID,
				Details:     "RDS instance is publicly accessible from the internet",
				Severity:    "critical",
				Remediation: "Disable public accessibility and use VPC security groups",
			})
		}
		if !rds.Encrypted {
			risks = append(risks, Risk{
				Type:        "Unencrypted RDS Instance",
				Resource:    rds.Name,
				ResourceID:  rds.ID,
				Details:     "RDS instance storage is not encrypted",
				Severity:    "high",
				Remediation: "Enable encryption at rest for the RDS instance",
			})
		}
	}

	// Check for instances with overly permissive IAM roles
	for _, inst := range resources.Instances {
		if inst.IAMRole != "" {
			// Check if the role contains admin or full access patterns
			if containsAdminPattern(inst.IAMRole) {
				risks = append(risks, Risk{
					Type:        "Overly Permissive IAM Role",
					Resource:    inst.Name,
					ResourceID:  inst.ID,
					Details:     fmt.Sprintf("Instance has potentially overly permissive IAM role: %s", inst.IAMRole),
					Severity:    "high",
					Remediation: "Apply principle of least privilege to IAM roles",
				})
			}
		}
	}

	return risks
}

func getPortName(port int32) string {
	switch port {
	case 22:
		return "SSH"
	case 3389:
		return "RDP"
	case 21:
		return "FTP"
	case 23:
		return "Telnet"
	case 0:
		return "All Ports"
	case 80:
		return "HTTP"
	case 443:
		return "HTTPS"
	case 3306:
		return "MySQL"
	case 5432:
		return "PostgreSQL"
	case 27017:
		return "MongoDB"
	case 6379:
		return "Redis"
	default:
		return "Unknown"
	}
}

func containsAdminPattern(role string) bool {
	patterns := []string{"Admin", "FullAccess", "PowerUser", "AdministratorAccess"}
	for _, pattern := range patterns {
		if strings.Contains(role, pattern) {
			return true
		}
	}
	return false
}
