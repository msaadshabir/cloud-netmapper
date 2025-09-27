# cloud-netmapper

A Go-based tool to automatically discover and visualize AWS VPC network topologies with basic security analysis.

## Features

- **Network Discovery**: Fetches VPCs, subnets, EC2 instances, security groups, and load balancers from AWS
- **Visual Mapping**: Generates network topology diagrams using Graphviz DOT format
- **Security Analysis**: Performs basic security risk assessment on discovered resources
- **Data Export**: Saves comprehensive resource data in JSON format
- **Command-line Tool**: Simple execution with clear console output

## Requirements

- Go 1.25.1 or later
- AWS CLI configured with appropriate IAM permissions
- Graphviz (`dot` command) for diagram generation

### AWS Permissions Required

The tool requires the following AWS permissions:

- `ec2:DescribeVpcs`
- `ec2:DescribeSubnets`
- `ec2:DescribeInstances`
- `ec2:DescribeSecurityGroups`
- `elasticloadbalancing:DescribeLoadBalancers`

## Installation

1. Clone the repository:

```bash
git clone https://github.com/saad-build/cloud-netmapper.git
cd cloud-netmapper
```

2. Install dependencies:

```bash
go mod tidy
```

3. Build the application:

```bash
go build
```

## Usage

1. Configure AWS credentials:

```bash
aws configure
```

2. Run the discovery (currently fixed to us-east-1):

```bash
./cloud-netmapper
```

Or run directly:

```bash
go run main.go
```

### Output

The tool generates:

- `aws_resources.json`: Complete AWS resource inventory
- `network_map.dot`: Graphviz source file
- `network_map.png`: Visual network diagram

### Console Output

```
Fetching AWS resources...
Raw data saved to aws_resources.json
Rendering diagram...
Diagram saved as network_map.png

SECURITY RISKS FOUND:
  • [High] Open Security Group: Port 22 open to 0.0.0.0/0 (Resource: default-sg)
```

## Security Analysis

Current security checks include:

- **Open Security Groups**: Ports 22, 3389, 21, 23, and 0 open to 0.0.0.0/0
- **Direct Public Exposure**: EC2 instances with public IPs but no load balancer protection

## Architecture

```
main.go                 # Application entry point and orchestration
├── aws_collector.go    # AWS API client and resource collection
├── visualizer.go       # Graphviz DOT file generation
├── security_checker.go # Security risk analysis
└── go.mod             # Go module dependencies
```

**"Failed to load AWS config"**

- Ensure AWS credentials are configured: `aws configure`
- Check IAM permissions include required EC2 and ELB permissions

**"Failed to render PNG"**

- Install Graphviz: `brew install graphviz` (macOS) or `sudo apt install graphviz` (Ubuntu)
- Ensure `dot` command is in PATH

**"no required module provides package"**

- Run `go mod tidy` to download dependencies
- Ensure Go 1.25.1+ is installed

### Debug Mode

For verbose output, modify the code to enable debug logging in the AWS SDK.

```bash
go build -o cloud-netmapper
```

### Running Tests

```bash
go test ./...
```

### Code Quality

```bash
golangci-lint run
```
