// Package pipeline implements the stage-based processing (Extract → Parse → Validate) of hashfm blocks.
package pipeline

import (
	"fmt"
  "strconv"
	"strings"

	"github.com/sidisinsane/hashfm"
	"github.com/sidisinsane/hashfm-agent/internal/model"
	"github.com/sidisinsane/hashfm-agent/internal/schema"
	"gopkg.in/yaml.v3"
)

// ErrNoBlock is returned by Process when a source file does not contain a valid hashfm delimiter.
type ErrNoBlock struct{}

func (ErrNoBlock) Error() string { return "no hashfm block found" }

// ErrInvalidBlock is returned by Process when a hashfm block fails schema validation.
type ErrInvalidBlock struct{ Reason string }

func (e ErrInvalidBlock) Error() string { return fmt.Sprintf("invalid hashfm: %s", e.Reason) }

// Process extracts, parses, and validates the hashfm-agent block from the provided source string.
// It returns a populated model.Block or an error if the block is missing (ErrNoBlock) or malformed.
func Process(src string) (model.Block, error) {
	rawYAML, err := hashfm.Extract(src)
	if err != nil {
		return model.Block{}, fmt.Errorf("extract: %w", err)
	}
	if rawYAML == "" {
		return model.Block{}, ErrNoBlock{}
	}

	block, err := parse(rawYAML)
	if err != nil {
		return model.Block{}, fmt.Errorf("parse: %w", err)
	}

	if err := validate(block); err != nil {
		return model.Block{}, fmt.Errorf("validate: %w", err)
	}

	return block, nil
}

// parse unmarshals raw YAML into a Block. Top-level sequence → multi-command;
// top-level mapping → single-command.
func parse(rawYAML string) (model.Block, error) {
	var top any
	if err := yaml.Unmarshal([]byte(rawYAML), &top); err != nil {
		return model.Block{}, err
	}

	switch top.(type) {
	case []any:
		var cmds []model.Command
		if err := yaml.Unmarshal([]byte(rawYAML), &cmds); err != nil {
			return model.Block{}, err
		}
		return model.Block{Multi: cmds}, nil

	case map[string]any:
		var cmd model.Command
		if err := yaml.Unmarshal([]byte(rawYAML), &cmd); err != nil {
			return model.Block{}, err
		}
		return model.Block{Single: &cmd}, nil

	default:
		return model.Block{}, fmt.Errorf("unexpected YAML top-level type: %T", top)
	}
}

// validate checks that all mandatory fields are present, then validates the block
// against the formal JSON schema. Returns ErrInvalidBlock on schema violations.
func validate(block model.Block) error {
	if block.IsMulti() {
		if len(block.Multi) < 2 {
			return fmt.Errorf("multi-command block must have at least 2 entries, got %d", len(block.Multi))
		}
		for i, cmd := range block.Multi {
			if err := validateCommand(cmd); err != nil {
				return fmt.Errorf("entry %d: %w", i, err)
			}
		}
		if err := schema.ValidateBlock(block.Multi); err != nil {
			return ErrInvalidBlock{Reason: err.Error()}
		}
		return nil
	}
	if err := validateCommand(*block.Single); err != nil {
		return err
	}
	if err := schema.ValidateBlock(block.Single); err != nil {
		return ErrInvalidBlock{Reason: err.Error()}
	}
	return nil
}

func validateCommand(cmd model.Command) error {
	if strings.TrimSpace(cmd.Description) == "" {
		return fmt.Errorf("missing required field: description")
	}
	if strings.TrimSpace(cmd.Usage) == "" {
		return fmt.Errorf("missing required field: usage")
	}
	if len(cmd.Exits) == 0 {
		return fmt.Errorf("missing required field: exits (must have at least one entry)")
	}
	for k := range cmd.Exits {
		if _, err := strconv.Atoi(k); err != nil {
			return fmt.Errorf("exits key %q must be an integer", k)
		}
	}
	return nil
}
