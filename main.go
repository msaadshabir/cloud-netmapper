package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("☁️  Fetching AWS resources...")
	resources, err := getAWSResources("us-east-1")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	rawData, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	err = os.WriteFile("aws_resources.json", rawData, 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Println("💾 Raw data saved to aws_resources.json")

	dotFile := "network_map.dot"
	err = generateDOTFile(resources, dotFile)
	if err != nil {
		log.Fatalf("Failed to generate DOT file: %v", err)
	}

	fmt.Println("🎨 Rendering diagram...")
	cmd := exec.Command("dot", "-Tpng", dotFile, "-o", "network_map.png")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to render PNG: %v", err)
	}
	fmt.Println("✅ Diagram saved as network_map.png")

	risks := checkSecurityRisks(resources)
	if len(risks) > 0 {
		fmt.Println("\n🚨 SECURITY RISKS FOUND:")
		for _, risk := range risks {
			fmt.Printf("  • [%s] %s: %s (Resource: %s)\n",
				risk.Severity, risk.Type, risk.Details, risk.Resource)
		}
	} else {
		fmt.Println("\n✅ No critical security risks detected.")
	}
}
