// Package index provides logic to transform a parsed block into high-level index entries.
package index

import (
	"path/filepath"
	"strings"

	"github.com/sidisinsane/hashfm-agent/internal/model"
)

// FromBlock derives a slice of IndexEntry from a parsed Block and its file path.
// The path should be relative to the workspace root to ensure portable indices.
func FromBlock(block model.Block, path string) []model.IndexEntry {
	scriptName := scriptNameFromPath(path)

	if !block.IsMulti() {
		return []model.IndexEntry{
			{
				Name:        scriptName,
				Path:        path,
				Description: block.Single.Description,
			},
		}
	}

	entries := make([]model.IndexEntry, 0, len(block.Multi))
	for _, cmd := range block.Multi {
		subcommand := subcommandFromUsage(cmd.Usage)
		name := scriptName
		if subcommand != "" {
			name = scriptName + " " + subcommand
		}
		entries = append(entries, model.IndexEntry{
			Name:        name,
			Path:        path,
			Description: cmd.Description,
		})
	}
	return entries
}

// scriptNameFromPath returns the filename without extension.
func scriptNameFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// subcommandFromUsage extracts the subcommand name — the second whitespace-delimited
// word of the usage string.
//
//	"git-tool.sh feature <branch-name>" → "feature"
func subcommandFromUsage(usage string) string {
	parts := strings.Fields(usage)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
