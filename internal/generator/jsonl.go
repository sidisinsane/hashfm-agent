// Package generator defines the interfaces and implementations for various index output formats.
package generator

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/sidisinsane/hashfm-agent/internal/model"
)

// JSONL implements the Generator interface to output one JSON object per line.
// This format is suitable for stream processing or large indices.
type JSONL struct{}

// Generate writes the index entries to the provided writer in JSONL format.
func (JSONL) Generate(w io.Writer, entries []model.IndexEntry) error {
    for _, e := range entries {
        // This local struct enforces the field order
        orderedEntry := struct {
            Name        string `json:"name"`
            Path        string `json:"path"`
            Description string `json:"description"`
        }{
            Name:        e.Name,
            Path:        e.Path,
            Description: e.Description,
        }

        b, err := json.Marshal(orderedEntry)
        if err != nil {
            return err
        }
        fmt.Fprintf(w, "%s\n", b)
    }
    return nil
}
