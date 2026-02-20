package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// FlowConfig represents the global flow configuration file.
type FlowConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// DefaultFlowConfig returns a FlowConfig with default values.
func DefaultFlowConfig() *FlowConfig {
	return &FlowConfig{
		APIVersion: "flow/v1",
		Kind:       "Config",
	}
}

// LoadFlowConfig reads and parses a flow config file from disk.
func LoadFlowConfig(path string) (*FlowConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var fc FlowConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &fc, nil
}

// SaveFlowConfig writes a FlowConfig to disk as YAML.
func SaveFlowConfig(path string, fc *FlowConfig) error {
	data, err := yaml.Marshal(fc)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}
