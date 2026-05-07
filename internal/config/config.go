// Package config handles the loading and validation of the .hashfm configuration file.
// It enforces namespaced settings to ensure compatibility with the hashfm convention.
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sidisinsane/hashfm-agent/internal/schema"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// Config represents the top-level .hashfm settings.
type Config struct {
	// Agent encapsulates all settings specific to the hashfm-agent implementation.
	Agent AgentConfig `yaml:"hashfm-agent"`
}

// AgentConfig represents the 'hashfm-agent' namespace in the configuration file.
// It groups settings related to agent-specific operations.
type AgentConfig struct {
	// Generate holds settings for the 'generate' subcommand.
	Generate GenerateConfig `yaml:"generate"`
}

// GenerateConfig defines the configuration options for index generation.
type GenerateConfig struct {
	// Format specifies the output format (e.g., "tsv", "jsonl", "yaml", "ndjson").
	Format string `yaml:"format"`
	// Output is the file path where the index will be written.
	Output string `yaml:"output"`
	// Recursive determines if subdirectories should be scanned for scripts.
	Recursive bool `yaml:"recursive"`
	// Include is a list of glob patterns to include. Applied before exclude.
	Include []string `yaml:"include"`
	// Exclude is a list of glob patterns to exclude.
	Exclude []string `yaml:"exclude"`
}

// Load reads and validates a .hashfm file from the provided path.
// If the file does not exist, it returns an empty Config and no error,
// allowing the application to proceed with default values.
func Load(configPath string) (*Config, error) {
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

	// Validate 'hashfm-agent' block against its schema if the block exists.
	if agentBlock, ok := raw["hashfm-agent"]; ok {
		if err := validateAgentConfig(agentBlock); err != nil {
			return nil, fmt.Errorf("config validation failed for 'hashfm-agent': %w", err)
		}
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config into struct: %w", err)
	}

	return &cfg, nil
}

// validateAgentConfig validates the agent-specific configuration block against
// the formal JSON schema to catch type mismatches or invalid options early.
func validateAgentConfig(agentBlock interface{}) error {
	// Use the embedded schema data.
	schemaData := schema.ConfigSchema

	// Since yaml unmarshals into map[string]interface{}, and gojsonschema wants JSON,
	// we convert the block to JSON for validation.
	jsonBlock, err := json.Marshal(agentBlock)
	if err != nil {
		return fmt.Errorf("failed to encode config block for validation: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
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
