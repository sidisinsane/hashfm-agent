// Package generator defines the interfaces and common types for all index representation formats.
package generator

import (
	"io"

	"github.com/sidisinsane/hashfm-agent/internal/model"
)

// Generator defines the standard interface for writing a collection of index entries to an output stream.
// All supported formats (TSV, JSONL, YAML) must implement this interface.
type Generator interface {
	// Generate formats and writes the provided entries to the writer.
	Generate(w io.Writer, entries []model.IndexEntry) error
}
