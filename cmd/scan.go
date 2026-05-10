package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sidisinsane/hashfm-agent/internal/generator"
	"github.com/sidisinsane/hashfm-agent/internal/index"
	"github.com/sidisinsane/hashfm-agent/internal/model"
	"github.com/sidisinsane/hashfm-agent/internal/pipeline"
)

// NewGenerator returns a concrete implementation of the Generator interface based on the requested format.
// Valid formats are "tsv", "jsonl", and "yaml". An empty string defaults to "tsv".
func NewGenerator(format string) (generator.Generator, error) {
	switch format {
	case "tsv", "":
		return generator.TSV{}, nil
	case "jsonl":
		return generator.JSONL{}, nil
	case "yaml":
		return generator.YAML{}, nil
	default:
		return nil, fmt.Errorf("unknown format %q: must be tsv, jsonl, or yaml", format)
	}
}

// ScanDir recursively or shallowly walks a directory to find and process matching .sh files.
// It returns a flattened list of all index entries found across all processed files.
func ScanDir(dir string, recursive bool, include, exclude []string) (entries []model.IndexEntry, warnings []string, err error) {
	walkFn := func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != dir && !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".sh" {
			return nil
		}

		// Compute rel for filtering — relative to dir, using forward slashes.
		filterRel, err := filepath.Rel(dir, path)
		if err != nil {
			filterRel = filepath.Base(path)
		}
		filterRel = filepath.ToSlash(filterRel)

		if len(include) > 0 && !matchesAny(filterRel, include) {
			return nil
		}
		if len(exclude) > 0 && matchesAny(filterRel, exclude) {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		block, err := pipeline.Process(string(src))
		if err != nil {
			if _, ok := err.(pipeline.ErrNoBlock); ok {
				return nil
			}
			warnings = append(warnings, fmt.Sprintf("%s: malformed hashfm block", path))
			return nil
		}

		rel, err := filepath.Rel(filepath.Dir(dir), path)
		if err != nil {
			rel = path
		}
		rel = "./" + filepath.ToSlash(rel)

		entries = append(entries, index.FromBlock(block, rel)...)
		return nil
	}

	err = filepath.WalkDir(dir, walkFn)
	return
}

func matchesAny(relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, relPath)
		if err == nil && matched {
			return true
		}
	}
	return false
}