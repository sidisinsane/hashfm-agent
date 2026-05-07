package generator

import (
	"fmt"
	"io"
	"strings"

	"github.com/sidisinsane/hashfm-agent/internal/model"
)

// YAML implements the Generator interface to output a YAML list of entries.
// This is useful for integration with other YAML-based automation tools.
type YAML struct{}

// Generate writes the index entries to the provided writer in a simplified YAML list format.
func (YAML) Generate(w io.Writer, entries []model.IndexEntry) error {
	for _, e := range entries {
		fmt.Fprintf(w, "- name: %s\n", yamlString(e.Name))
		fmt.Fprintf(w, "  path: %s\n", yamlString(e.Path))
		fmt.Fprintf(w, "  description: %s\n", yamlString(e.Description))
	}
	return nil
}

// yamlString quotes a string if it contains spaces or special YAML characters.
func yamlString(s string) string {
	if strings.ContainsAny(s, " \t:#{}[]|>&*!,\"'\\") {
		// Simple double-quote with escaping
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return s
}
