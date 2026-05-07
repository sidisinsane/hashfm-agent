// Package schema provides embedded JSON schemas for configuration validation.
package schema

import _ "embed"

// ConfigSchema is the embedded schema for the .hashfm config file (hashfm-agent namespace).
//go:embed hashfm-agent-config.schema.json
var ConfigSchema []byte

// BlockSchema is the embedded schema for script hashfms.
//go:embed hashfm-agent.schema.json
var BlockSchema []byte

//go:generate go run generate.go
