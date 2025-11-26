package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Regions  []string       `yaml:"regions"`
	Output   OutputConfig   `yaml:"output"`
	Security SecurityConfig `yaml:"security"`
}

// OutputConfig contains output-related settings
type OutputConfig struct {
	Format    string `yaml:"format"`    // png, svg, html, json, csv, markdown, sarif
	Directory string `yaml:"directory"` // output directory
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	SeverityThreshold string `yaml:"severity_threshold"` // low, medium, high, critical
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Regions: []string{"us-east-1"},
		Output: OutputConfig{
			Format:    "png",
			Directory: ".",
		},
		Security: SecurityConfig{
			SeverityThreshold: "low",
		},
	}
}

// LoadConfig loads configuration from file or returns defaults
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// Try to find config file
	if path == "" {
		// Look for config in current directory and home directory
		candidates := []string{
			".cloud-netmapper.yaml",
			".cloud-netmapper.yml",
			filepath.Join(os.Getenv("HOME"), ".cloud-netmapper.yaml"),
			filepath.Join(os.Getenv("HOME"), ".cloud-netmapper.yml"),
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				path = candidate
				break
			}
		}
	}

	if path == "" {
		return config, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
