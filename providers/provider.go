package providers

import (
	"fmt"
	"strings"

	"cloud-netmapper/types"
)

// CloudProvider is the interface for cloud providers
type CloudProvider interface {
	// Name returns the provider name
	Name() string

	// GetResources retrieves all resources from the specified region
	GetResources(region string) (*types.CloudResources, error)

	// ListRegions returns all available regions
	ListRegions() ([]string, error)
}

// SupportedProviders lists all supported cloud providers
var SupportedProviders = []string{"aws", "azure", "gcp"}

// GetProvider returns a CloudProvider instance for the specified provider name
func GetProvider(name string) (CloudProvider, error) {
	switch strings.ToLower(name) {
	case "aws":
		return NewAWSProvider()
	case "azure":
		// Azure requires subscription ID from environment or config
		// For now, return a placeholder error
		return nil, fmt.Errorf("azure provider requires subscription ID - use NewAzureProvider() directly")
	case "gcp":
		// GCP requires project ID from environment or config
		// For now, return a placeholder error
		return nil, fmt.Errorf("gcp provider requires project ID - use NewGCPProvider() directly")
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s (supported: %v)", name, SupportedProviders)
	}
}
