package generator

import (
	"fmt"
	"io"

	"github.com/sidisinsane/hashfm-agent/internal/model"
)

// TSV implements the Generator interface to output a tab-separated index.
// This format is highly readable and easy to parse with standard Unix tools.
type TSV struct{}

// Generate writes the index entries to the provided writer in TSV format with a header line.
func (TSV) Generate(w io.Writer, entries []model.IndexEntry) error {
	fmt.Fprintln(w, "name\tpath\tdescription")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.Name, e.Path, e.Description)
	}
	return nil
}
