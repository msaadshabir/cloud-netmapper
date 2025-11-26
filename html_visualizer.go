package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Cloud NetMapper - Network Visualization</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/cytoscape/3.28.1/cytoscape.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: #1a1a2e;
            color: #eee;
        }
        .container {
            display: flex;
            height: 100vh;
        }
        .sidebar {
            width: 300px;
            background: #16213e;
            padding: 20px;
            overflow-y: auto;
            border-right: 1px solid #0f3460;
        }
        .sidebar h1 {
            font-size: 1.5rem;
            margin-bottom: 20px;
            color: #e94560;
        }
        .sidebar h2 {
            font-size: 1rem;
            margin: 15px 0 10px;
            color: #0f4c75;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .filter-group {
            margin-bottom: 15px;
        }
        .filter-group label {
            display: flex;
            align-items: center;
            padding: 5px 0;
            cursor: pointer;
        }
        .filter-group input[type="checkbox"] {
            margin-right: 10px;
        }
        .search-box {
            width: 100%;
            padding: 10px;
            border: 1px solid #0f3460;
            border-radius: 5px;
            background: #1a1a2e;
            color: #eee;
            margin-bottom: 15px;
        }
        .search-box:focus {
            outline: none;
            border-color: #e94560;
        }
        .stats {
            background: #0f3460;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .stats-item {
            display: flex;
            justify-content: space-between;
            padding: 5px 0;
            border-bottom: 1px solid #16213e;
        }
        .stats-item:last-child {
            border-bottom: none;
        }
        .stats-value {
            font-weight: bold;
            color: #e94560;
        }
        #cy {
            flex: 1;
            background: #1a1a2e;
        }
        .node-details {
            position: fixed;
            bottom: 20px;
            right: 20px;
            width: 350px;
            background: #16213e;
            border: 1px solid #0f3460;
            border-radius: 8px;
            padding: 20px;
            display: none;
            max-height: 400px;
            overflow-y: auto;
        }
        .node-details.active {
            display: block;
        }
        .node-details h3 {
            color: #e94560;
            margin-bottom: 15px;
        }
        .node-details .close-btn {
            position: absolute;
            top: 10px;
            right: 15px;
            cursor: pointer;
            font-size: 1.5rem;
            color: #888;
        }
        .node-details .close-btn:hover {
            color: #e94560;
        }
        .detail-row {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #0f3460;
        }
        .detail-row:last-child {
            border-bottom: none;
        }
        .detail-label {
            color: #888;
        }
        .detail-value {
            color: #eee;
            text-align: right;
            word-break: break-all;
        }
        .legend {
            margin-top: 20px;
        }
        .legend-item {
            display: flex;
            align-items: center;
            margin: 5px 0;
        }
        .legend-color {
            width: 20px;
            height: 20px;
            border-radius: 50%;
            margin-right: 10px;
        }
        .controls {
            margin-top: 20px;
        }
        .btn {
            width: 100%;
            padding: 10px;
            margin: 5px 0;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 0.9rem;
        }
        .btn-primary {
            background: #e94560;
            color: white;
        }
        .btn-secondary {
            background: #0f3460;
            color: white;
        }
        .btn:hover {
            opacity: 0.9;
        }
        .risk-badge {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 0.8rem;
            margin-left: 5px;
        }
        .risk-critical { background: #dc3545; }
        .risk-high { background: #fd7e14; }
        .risk-medium { background: #ffc107; color: #000; }
        .risk-low { background: #17a2b8; }
    </style>
</head>
<body>
    <div class="container">
        <div class="sidebar">
            <h1>Cloud NetMapper</h1>
            
            <input type="text" class="search-box" id="searchBox" placeholder="Search resources...">
            
            <div class="stats">
                <div class="stats-item">
                    <span>VPCs</span>
                    <span class="stats-value">{{.Stats.VPCs}}</span>
                </div>
                <div class="stats-item">
                    <span>Subnets</span>
                    <span class="stats-value">{{.Stats.Subnets}}</span>
                </div>
                <div class="stats-item">
                    <span>EC2 Instances</span>
                    <span class="stats-value">{{.Stats.Instances}}</span>
                </div>
                <div class="stats-item">
                    <span>Security Groups</span>
                    <span class="stats-value">{{.Stats.SecurityGroups}}</span>
                </div>
                <div class="stats-item">
                    <span>Load Balancers</span>
                    <span class="stats-value">{{.Stats.LoadBalancers}}</span>
                </div>
                <div class="stats-item">
                    <span>RDS Instances</span>
                    <span class="stats-value">{{.Stats.RDSInstances}}</span>
                </div>
                <div class="stats-item">
                    <span>NAT Gateways</span>
                    <span class="stats-value">{{.Stats.NATGateways}}</span>
                </div>
                <div class="stats-item">
                    <span>Lambda Functions</span>
                    <span class="stats-value">{{.Stats.LambdaFunctions}}</span>
                </div>
            </div>

            <h2>Filter by Type</h2>
            <div class="filter-group">
                <label><input type="checkbox" class="filter-checkbox" data-type="vpc" checked> VPCs</label>
                <label><input type="checkbox" class="filter-checkbox" data-type="subnet" checked> Subnets</label>
                <label><input type="checkbox" class="filter-checkbox" data-type="instance" checked> EC2 Instances</label>
                <label><input type="checkbox" class="filter-checkbox" data-type="lb" checked> Load Balancers</label>
                <label><input type="checkbox" class="filter-checkbox" data-type="rds" checked> RDS</label>
                <label><input type="checkbox" class="filter-checkbox" data-type="nat" checked> NAT Gateways</label>
                <label><input type="checkbox" class="filter-checkbox" data-type="igw" checked> Internet Gateways</label>
                <label><input type="checkbox" class="filter-checkbox" data-type="lambda" checked> Lambda</label>
                <label><input type="checkbox" class="filter-checkbox" data-type="eks" checked> EKS</label>
                <label><input type="checkbox" class="filter-checkbox" data-type="ecs" checked> ECS</label>
            </div>

            <h2>Legend</h2>
            <div class="legend">
                <div class="legend-item"><div class="legend-color" style="background: #4a90d9;"></div>VPC</div>
                <div class="legend-item"><div class="legend-color" style="background: #f5a623;"></div>Subnet</div>
                <div class="legend-item"><div class="legend-color" style="background: #7ed321;"></div>EC2 Instance</div>
                <div class="legend-item"><div class="legend-color" style="background: #bd10e0;"></div>Load Balancer</div>
                <div class="legend-item"><div class="legend-color" style="background: #50e3c2;"></div>NAT Gateway</div>
                <div class="legend-item"><div class="legend-color" style="background: #9013fe;"></div>Internet Gateway</div>
                <div class="legend-item"><div class="legend-color" style="background: #d0021b;"></div>RDS</div>
                <div class="legend-item"><div class="legend-color" style="background: #ff6b35;"></div>Lambda</div>
                <div class="legend-item"><div class="legend-color" style="background: #00bcd4;"></div>EKS</div>
                <div class="legend-item"><div class="legend-color" style="background: #ff9800;"></div>ECS</div>
            </div>

            <div class="controls">
                <button class="btn btn-primary" onclick="resetLayout()">Reset Layout</button>
                <button class="btn btn-secondary" onclick="fitToScreen()">Fit to Screen</button>
                <button class="btn btn-secondary" onclick="exportPNG()">Export PNG</button>
            </div>
        </div>
        
        <div id="cy"></div>
        
        <div class="node-details" id="nodeDetails">
            <span class="close-btn" onclick="closeDetails()">&times;</span>
            <h3 id="detailsTitle">Resource Details</h3>
            <div id="detailsContent"></div>
        </div>
    </div>

    <script>
        const graphData = {{.GraphData}};
        
        const cy = cytoscape({
            container: document.getElementById('cy'),
            elements: graphData,
            style: [
                {
                    selector: 'node',
                    style: {
                        'label': 'data(label)',
                        'text-valign': 'bottom',
                        'text-halign': 'center',
                        'font-size': '10px',
                        'color': '#eee',
                        'text-margin-y': 5,
                        'width': 40,
                        'height': 40
                    }
                },
                {
                    selector: 'node[type="vpc"]',
                    style: {
                        'background-color': '#4a90d9',
                        'shape': 'rectangle',
                        'width': 60,
                        'height': 40
                    }
                },
                {
                    selector: 'node[type="subnet"]',
                    style: {
                        'background-color': '#f5a623',
                        'shape': 'ellipse'
                    }
                },
                {
                    selector: 'node[type="instance"]',
                    style: {
                        'background-color': '#7ed321',
                        'shape': 'ellipse'
                    }
                },
                {
                    selector: 'node[type="lb"]',
                    style: {
                        'background-color': '#bd10e0',
                        'shape': 'diamond',
                        'width': 45,
                        'height': 45
                    }
                },
                {
                    selector: 'node[type="nat"]',
                    style: {
                        'background-color': '#50e3c2',
                        'shape': 'triangle'
                    }
                },
                {
                    selector: 'node[type="igw"]',
                    style: {
                        'background-color': '#9013fe',
                        'shape': 'star',
                        'width': 50,
                        'height': 50
                    }
                },
                {
                    selector: 'node[type="rds"]',
                    style: {
                        'background-color': '#d0021b',
                        'shape': 'barrel'
                    }
                },
                {
                    selector: 'node[type="lambda"]',
                    style: {
                        'background-color': '#ff6b35',
                        'shape': 'hexagon'
                    }
                },
                {
                    selector: 'node[type="eks"]',
                    style: {
                        'background-color': '#00bcd4',
                        'shape': 'octagon'
                    }
                },
                {
                    selector: 'node[type="ecs"]',
                    style: {
                        'background-color': '#ff9800',
                        'shape': 'pentagon'
                    }
                },
                {
                    selector: 'edge',
                    style: {
                        'width': 2,
                        'line-color': '#555',
                        'target-arrow-color': '#555',
                        'target-arrow-shape': 'triangle',
                        'curve-style': 'bezier'
                    }
                },
                {
                    selector: '.highlighted',
                    style: {
                        'border-width': 3,
                        'border-color': '#e94560'
                    }
                },
                {
                    selector: '.faded',
                    style: {
                        'opacity': 0.2
                    }
                }
            ],
            layout: {
                name: 'cose',
                idealEdgeLength: 100,
                nodeOverlap: 20,
                refresh: 20,
                fit: true,
                padding: 30,
                randomize: false,
                componentSpacing: 100,
                nodeRepulsion: 400000,
                edgeElasticity: 100,
                nestingFactor: 5,
                gravity: 80,
                numIter: 1000,
                initialTemp: 200,
                coolingFactor: 0.95,
                minTemp: 1.0
            }
        });

        // Node click handler
        cy.on('tap', 'node', function(evt) {
            const node = evt.target;
            showDetails(node.data());
        });

        // Background click handler
        cy.on('tap', function(evt) {
            if (evt.target === cy) {
                closeDetails();
            }
        });

        function showDetails(data) {
            const details = document.getElementById('nodeDetails');
            const title = document.getElementById('detailsTitle');
            const content = document.getElementById('detailsContent');
            
            title.textContent = data.label || data.id;
            
            let html = '';
            for (const [key, value] of Object.entries(data)) {
                if (key !== 'id' && key !== 'label' && key !== 'type' && value) {
                    const displayKey = key.replace(/([A-Z])/g, ' $1').replace(/^./, str => str.toUpperCase());
                    let displayValue = value;
                    if (Array.isArray(value)) {
                        displayValue = value.join(', ') || 'None';
                    }
                    html += '<div class="detail-row"><span class="detail-label">' + displayKey + '</span><span class="detail-value">' + displayValue + '</span></div>';
                }
            }
            
            content.innerHTML = html || '<p>No additional details available.</p>';
            details.classList.add('active');
        }

        function closeDetails() {
            document.getElementById('nodeDetails').classList.remove('active');
        }

        // Search functionality
        document.getElementById('searchBox').addEventListener('input', function(e) {
            const query = e.target.value.toLowerCase();
            
            if (query === '') {
                cy.elements().removeClass('faded highlighted');
                return;
            }
            
            cy.elements().addClass('faded');
            cy.nodes().filter(node => {
                const label = (node.data('label') || '').toLowerCase();
                const id = (node.data('id') || '').toLowerCase();
                return label.includes(query) || id.includes(query);
            }).removeClass('faded').addClass('highlighted');
        });

        // Filter functionality
        document.querySelectorAll('.filter-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', function() {
                const type = this.dataset.type;
                const visible = this.checked;
                
                cy.nodes('[type="' + type + '"]').style('display', visible ? 'element' : 'none');
            });
        });

        function resetLayout() {
            cy.layout({
                name: 'cose',
                animate: true,
                animationDuration: 500
            }).run();
        }

        function fitToScreen() {
            cy.fit(cy.elements(), 50);
        }

        function exportPNG() {
            const png = cy.png({scale: 2, bg: '#1a1a2e'});
            const link = document.createElement('a');
            link.download = 'network-map.png';
            link.href = png;
            link.click();
        }
    </script>
</body>
</html>`

type HTMLStats struct {
	VPCs            int
	Subnets         int
	Instances       int
	SecurityGroups  int
	LoadBalancers   int
	RDSInstances    int
	NATGateways     int
	LambdaFunctions int
}

type HTMLData struct {
	Stats     HTMLStats
	GraphData string
}

type CyNode struct {
	Data map[string]interface{} `json:"data"`
}

type CyEdge struct {
	Data map[string]interface{} `json:"data"`
}

type CyElements struct {
	Nodes []CyNode `json:"nodes"`
	Edges []CyEdge `json:"edges"`
}

func generateHTMLFile(resources *AWSResources, filename string) error {
	elements := CyElements{
		Nodes: []CyNode{},
		Edges: []CyEdge{},
	}

	// Add VPCs
	for _, vpc := range resources.VPCs {
		nodeID := "vpc_" + sanitizeID(vpc.ID)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":              nodeID,
				"label":           truncateLabel(vpc.Name, 20),
				"type":            "vpc",
				"vpcId":           vpc.ID,
				"cidr":            vpc.CIDR,
				"isDefault":       vpc.IsDefault,
				"flowLogsEnabled": vpc.FlowLogsEnabled,
			},
		})
	}

	// Add Subnets
	for _, subnet := range resources.Subnets {
		nodeID := "subnet_" + sanitizeID(subnet.ID)
		vpcID := "vpc_" + sanitizeID(subnet.VPCID)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":       nodeID,
				"label":    truncateLabel(subnet.Name, 15),
				"type":     "subnet",
				"subnetId": subnet.ID,
				"cidr":     subnet.CIDR,
				"az":       subnet.AZ,
				"isPublic": subnet.IsPublic,
			},
		})
		elements.Edges = append(elements.Edges, CyEdge{
			Data: map[string]interface{}{
				"id":     fmt.Sprintf("edge_%s_%s", vpcID, nodeID),
				"source": vpcID,
				"target": nodeID,
			},
		})
	}

	// Add Instances
	for _, inst := range resources.Instances {
		nodeID := "instance_" + sanitizeID(inst.ID)
		subnetID := "subnet_" + sanitizeID(inst.SubnetID)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":           nodeID,
				"label":        truncateLabel(inst.Name, 15),
				"type":         "instance",
				"instanceId":   inst.ID,
				"privateIp":    inst.PrivateIP,
				"publicIp":     inst.PublicIP,
				"instanceType": inst.InstanceType,
			},
		})
		elements.Edges = append(elements.Edges, CyEdge{
			Data: map[string]interface{}{
				"id":     fmt.Sprintf("edge_%s_%s", subnetID, nodeID),
				"source": subnetID,
				"target": nodeID,
			},
		})
	}

	// Add Load Balancers
	for _, lb := range resources.LoadBalancers {
		nodeID := "lb_" + sanitizeID(lb.ARN)
		vpcID := "vpc_" + sanitizeID(lb.VPCID)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":     nodeID,
				"label":  truncateLabel(lb.Name, 15),
				"type":   "lb",
				"arn":    lb.ARN,
				"scheme": lb.Scheme,
				"lbType": lb.Type,
			},
		})
		elements.Edges = append(elements.Edges, CyEdge{
			Data: map[string]interface{}{
				"id":     fmt.Sprintf("edge_%s_%s", vpcID, nodeID),
				"source": vpcID,
				"target": nodeID,
			},
		})
	}

	// Add NAT Gateways
	for _, nat := range resources.NATGateways {
		nodeID := "nat_" + sanitizeID(nat.ID)
		subnetID := "subnet_" + sanitizeID(nat.SubnetID)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":        nodeID,
				"label":     truncateLabel(nat.Name, 15),
				"type":      "nat",
				"natId":     nat.ID,
				"publicIp":  nat.PublicIP,
				"privateIp": nat.PrivateIP,
				"state":     nat.State,
			},
		})
		if nat.SubnetID != "" {
			elements.Edges = append(elements.Edges, CyEdge{
				Data: map[string]interface{}{
					"id":     fmt.Sprintf("edge_%s_%s", subnetID, nodeID),
					"source": subnetID,
					"target": nodeID,
				},
			})
		}
	}

	// Add Internet Gateways
	for _, igw := range resources.InternetGateways {
		nodeID := "igw_" + sanitizeID(igw.ID)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":    nodeID,
				"label": truncateLabel(igw.Name, 15),
				"type":  "igw",
				"igwId": igw.ID,
			},
		})
		for _, vpcID := range igw.VPCIDs {
			vpcNodeID := "vpc_" + sanitizeID(vpcID)
			elements.Edges = append(elements.Edges, CyEdge{
				Data: map[string]interface{}{
					"id":     fmt.Sprintf("edge_%s_%s", vpcNodeID, nodeID),
					"source": vpcNodeID,
					"target": nodeID,
				},
			})
		}
	}

	// Add RDS Instances
	for _, rds := range resources.RDSInstances {
		nodeID := "rds_" + sanitizeID(rds.ID)
		vpcID := "vpc_" + sanitizeID(rds.VPCID)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":        nodeID,
				"label":     truncateLabel(rds.Name, 15),
				"type":      "rds",
				"rdsId":     rds.ID,
				"engine":    rds.Engine,
				"encrypted": rds.Encrypted,
				"public":    rds.PubliclyAccessible,
			},
		})
		if rds.VPCID != "" {
			elements.Edges = append(elements.Edges, CyEdge{
				Data: map[string]interface{}{
					"id":     fmt.Sprintf("edge_%s_%s", vpcID, nodeID),
					"source": vpcID,
					"target": nodeID,
				},
			})
		}
	}

	// Add Lambda Functions (only those with VPC config)
	for _, fn := range resources.LambdaFunctions {
		if fn.VPCID == "" {
			continue
		}
		nodeID := "lambda_" + sanitizeID(fn.Name)
		vpcID := "vpc_" + sanitizeID(fn.VPCID)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":      nodeID,
				"label":   truncateLabel(fn.Name, 15),
				"type":    "lambda",
				"arn":     fn.ARN,
				"runtime": fn.Runtime,
			},
		})
		elements.Edges = append(elements.Edges, CyEdge{
			Data: map[string]interface{}{
				"id":     fmt.Sprintf("edge_%s_%s", vpcID, nodeID),
				"source": vpcID,
				"target": nodeID,
			},
		})
	}

	// Add EKS Clusters
	for _, cluster := range resources.EKSClusters {
		nodeID := "eks_" + sanitizeID(cluster.Name)
		vpcID := "vpc_" + sanitizeID(cluster.VPCID)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":     nodeID,
				"label":  truncateLabel(cluster.Name, 15),
				"type":   "eks",
				"arn":    cluster.ARN,
				"status": cluster.Status,
			},
		})
		if cluster.VPCID != "" {
			elements.Edges = append(elements.Edges, CyEdge{
				Data: map[string]interface{}{
					"id":     fmt.Sprintf("edge_%s_%s", vpcID, nodeID),
					"source": vpcID,
					"target": nodeID,
				},
			})
		}
	}

	// Add ECS Clusters
	for _, cluster := range resources.ECSClusters {
		nodeID := "ecs_" + sanitizeID(cluster.Name)
		elements.Nodes = append(elements.Nodes, CyNode{
			Data: map[string]interface{}{
				"id":       nodeID,
				"label":    truncateLabel(cluster.Name, 15),
				"type":     "ecs",
				"arn":      cluster.ARN,
				"status":   cluster.Status,
				"services": cluster.ServiceCount,
				"tasks":    cluster.TaskCount,
			},
		})
	}

	// Convert elements to JSON
	graphJSON, err := json.Marshal(elements)
	if err != nil {
		return fmt.Errorf("failed to marshal graph data: %v", err)
	}

	// Prepare template data
	data := HTMLData{
		Stats: HTMLStats{
			VPCs:            len(resources.VPCs),
			Subnets:         len(resources.Subnets),
			Instances:       len(resources.Instances),
			SecurityGroups:  len(resources.SecurityGroups),
			LoadBalancers:   len(resources.LoadBalancers),
			RDSInstances:    len(resources.RDSInstances),
			NATGateways:     len(resources.NATGateways),
			LambdaFunctions: len(resources.LambdaFunctions),
		},
		GraphData: string(graphJSON),
	}

	// Parse and execute template
	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %v", err)
	}

	return nil
}

func sanitizeID(id string) string {
	replacer := strings.NewReplacer(
		":", "_",
		"/", "_",
		"-", "_",
		".", "_",
	)
	return replacer.Replace(id)
}

func truncateLabel(label string, maxLen int) string {
	if len(label) <= maxLen {
		return label
	}
	return label[:maxLen-3] + "..."
}
