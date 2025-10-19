# Cloud NetMapper

[![Go Version](https://img.shields.io/badge/Go-1.25.1+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![AWS](https://img.shields.io/badge/AWS-Cloud-FF9900?style=flat&logo=amazon-aws)](https://aws.amazon.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat)](LICENSE)
[![Graphviz](https://img.shields.io/badge/Graphviz-Required-5E5086?style=flat)](https://graphviz.org/)

> Visualize your AWS network topology and detect security risks automatically

Discover VPCs, subnets, instances, security groups, and load balancers—then generate a visual network diagram with built-in security analysis.

---

## Features

**Auto-Discovery** · Map your entire AWS network  
**Visual Diagrams** · Generate PNG topology maps  
**Security Scanning** · Detect open ports and misconfigurations  
**JSON Export** · Save complete resource inventory

---

## Quick Start

### Prerequisites

```bash
# macOS
brew install go graphviz

# Ubuntu/Debian
sudo apt install golang graphviz
```

> **Requirements:** Go 1.25.1+, AWS CLI configured, Graphviz

### Install

```bash
git clone https://github.com/msaadshabir/cloud-netmapper.git
cd cloud-netmapper
go build
```

### Run

```bash
aws configure                # Set up AWS credentials
./cloud-netmapper           # Scan us-east-1 region
```

---

## Output

| File                 | Description                     |
| -------------------- | ------------------------------- |
| `network_map.png`    | Visual network topology diagram |
| `network_map.dot`    | Graphviz source file            |
| `aws_resources.json` | Complete resource inventory     |

---

## Security Checks

- Open SSH/RDP/FTP ports to `0.0.0.0/0`
- EC2 instances exposed without load balancers
- Security group misconfigurations

---

## IAM Permissions

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeVpcs",
        "ec2:DescribeSubnets",
        "ec2:DescribeInstances",
        "ec2:DescribeSecurityGroups",
        "elasticloadbalancing:DescribeLoadBalancers"
      ],
      "Resource": "*"
    }
  ]
}
```

---

## Troubleshooting

**AWS credentials not found**  
→ Run `aws configure`

**Graphviz not installed**  
→ `brew install graphviz` or `sudo apt install graphviz`

**Module errors**  
→ Run `go mod tidy`

---

## License

MIT
