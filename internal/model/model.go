// Package model defines the core data structures used throughout the hashfm-agent.
package model

// Command represents a single command definition extracted from a hashfm block.
// It can represent either a standalone script or a single subcommand within a multi-command script.
type Command struct {
	Description string            `yaml:"description"`
	Usage       string            `yaml:"usage"`
	Exits       map[string]string `yaml:"exits"`
}

// Block represents a parsed hashfm-agent metadata block.
// To support both simple and complex scripts, exactly one of Single or Multi is populated.
type Block struct {
	Single *Command
	Multi  []Command
}

// IsMulti reports whether the block contains multiple subcommand definitions.
func (b Block) IsMulti() bool {
	return b.Multi != nil
}

// IndexEntry represents a single entry in the generated index.
// It maps a specific command or subcommand to its source script and description.
type IndexEntry struct {
	Name        string // "git-tool feature" or "convert-images"
	Path        string // relative path to script
	Description string // from the command entry
}
