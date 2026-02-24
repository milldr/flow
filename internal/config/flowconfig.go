package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Agent represents a configured agent tool (editor, AI assistant, etc.).
type Agent struct {
	Name    string `yaml:"name"`
	Exec    string `yaml:"exec"`
	Default bool   `yaml:"default,omitempty"`
}

// FlowConfigSpec holds optional configuration nested under spec.
type FlowConfigSpec struct {
	Agents []Agent `yaml:"agents,omitempty"`
}

// FlowConfig represents the global flow configuration file.
type FlowConfig struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Spec       FlowConfigSpec `yaml:"spec,omitempty"`
}

// DefaultAgent returns the agent marked as default, or nil if none is configured.
func (fc *FlowConfig) DefaultAgent() *Agent {
	for i := range fc.Spec.Agents {
		if fc.Spec.Agents[i].Default {
			return &fc.Spec.Agents[i]
		}
	}
	return nil
}

// DefaultFlowConfig returns a FlowConfig with default values.
func DefaultFlowConfig() *FlowConfig {
	return &FlowConfig{
		APIVersion: "flow/v1",
		Kind:       "Config",
		Spec: FlowConfigSpec{
			Agents: []Agent{
				{Name: "claude", Exec: "claude", Default: true},
			},
		},
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
