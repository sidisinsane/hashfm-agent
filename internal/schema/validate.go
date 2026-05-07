package schema

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// ValidateBlock validates a parsed hashfm block against the block schema.
// The block is either a single command (map) or multi-command (array).
// Returns nil if the block is valid. Returns a descriptive error if invalid.
// The caller decides how to handle the error — skip + warn, or fail.
func ValidateBlock(block interface{}) error {
	jsonBlock, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("encode block for validation: %w", err)
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(BlockSchema),
		gojsonschema.NewBytesLoader(jsonBlock),
	)
	if err != nil {
		return fmt.Errorf("validator error: %w", err)
	}

	if !result.Valid() {
		var errMsgs string
		for _, desc := range result.Errors() {
			errMsgs += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf("block schema violations:\n%s", errMsgs)
	}

	return nil
}
