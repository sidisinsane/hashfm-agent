// Package config handles loading and validating the hashfm-agent specific
// configuration from the shared .hashfm config file, using hashfm.LoadConfig.
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sidisinsane/hashfm"
	"github.com/sidisinsane/hashfm-agent/internal/schema"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// Config represents the hashfm-agent namespace in the .hashfm configuration.
type Config struct {
	Agent AgentConfig `yaml:"hashfm-agent"`
}

// AgentConfig represents the 'hashfm-agent' namespace in the configuration file.
type AgentConfig struct {
	Generate GenerateConfig `yaml:"generate"`
}

// GenerateConfig defines the configuration options for index generation.
type GenerateConfig struct {
	Format    string   `yaml:"format"`
	Output    string   `yaml:"output"`
	Recursive bool     `yaml:"recursive"`
	Include   []string `yaml:"include"`
	Exclude   []string `yaml:"exclude"`
}

// Load reads and validates the .hashfm config file using hashfm.LoadConfig.
// It extracts the 'hashfm-agent' namespace and validates it against the agent's schema.
// Returns a default Config if no config file is found.
func Load() (*Config, error) {
	rawConfig, err := hashfm.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if rawConfig == nil {
		return &Config{}, nil
	}

	agentBlock, ok := rawConfig["hashfm-agent"]
	if !ok {
		return &Config{}, nil
	}

	return extractAgentConfig(agentBlock)
}

// LoadWithPath loads a config from a specific path (for CLI override).
// This allows users to specify a different config file via the --config flag.
func LoadWithPath(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	agentBlock, ok := raw["hashfm-agent"]
	if !ok {
		return &Config{}, nil
	}

	return extractAgentConfig(agentBlock)
}

// extractAgentConfig validates and maps an agent config block into a Config struct.
// Shared by Load and LoadWithPath.
func extractAgentConfig(agentBlock interface{}) (*Config, error) {
	if err := validateAgentConfig(agentBlock); err != nil {
		return nil, fmt.Errorf("config validation failed for 'hashfm-agent': %w", err)
	}

	var cfg Config
	if gen, ok := agentBlock.(map[string]interface{})["generate"].(map[string]interface{}); ok {
		if format, ok := gen["format"].(string); ok {
			cfg.Agent.Generate.Format = format
		}
		if output, ok := gen["output"].(string); ok {
			cfg.Agent.Generate.Output = output
		}
		if recursive, ok := gen["recursive"].(bool); ok {
			cfg.Agent.Generate.Recursive = recursive
		}
		if include, ok := gen["include"].([]interface{}); ok {
			cfg.Agent.Generate.Include = toStringSlice(include)
		}
		if exclude, ok := gen["exclude"].([]interface{}); ok {
			cfg.Agent.Generate.Exclude = toStringSlice(exclude)
		}
	}

	return &cfg, nil
}

// validateAgentConfig validates the agent-specific configuration block against
// the formal JSON schema.
func validateAgentConfig(agentBlock interface{}) error {
	jsonBlock, err := json.Marshal(agentBlock)
	if err != nil {
		return fmt.Errorf("failed to encode config block for validation: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schema.ConfigSchema)
	documentLoader := gojsonschema.NewBytesLoader(jsonBlock)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validator internal error: %w", err)
	}

	if !result.Valid() {
		var errMsgs string
		for _, desc := range result.Errors() {
			errMsgs += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf("schema violations:\n%s", errMsgs)
	}

	return nil
}

func toStringSlice(in []interface{}) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = v.(string)
	}
	return out
}
