package cmd_test

import (
	"path/filepath"
	"testing"

	"github.com/sidisinsane/hashfm-agent/cmd"
	"github.com/sidisinsane/hashfm-agent/internal/generator"
)

const testdir = "../testdata"

func testPath(name string) string {
	return filepath.Join(testdir, name)
}

// NewGenerator tests

func TestNewGenerator_Defaults(t *testing.T) {
	for _, format := range []string{"tsv", ""} {
		gen, err := cmd.NewGenerator(format)
		if err != nil {
			t.Errorf("format %q: unexpected error: %v", format, err)
		}
		if _, ok := gen.(generator.TSV); !ok {
			t.Errorf("format %q: expected TSV generator", format)
		}
	}
}

func TestNewGenerator_AllFormats(t *testing.T) {
	cases := []struct {
		format string
		want   generator.Generator
	}{
		{"tsv", generator.TSV{}},
		{"jsonl", generator.JSONL{}},
		{"yaml", generator.YAML{}},
	}
	for _, tc := range cases {
		gen, err := cmd.NewGenerator(tc.format)
		if err != nil {
			t.Errorf("format %q: unexpected error: %v", tc.format, err)
		}
		if gen != tc.want {
			t.Errorf("format %q: got %T, want %T", tc.format, gen, tc.want)
		}
	}
}

func TestNewGenerator_UnknownFormat(t *testing.T) {
	_, err := cmd.NewGenerator("xml")
	if err == nil {
		t.Error("expected error for unknown format, got nil")
	}
}

// ScanDir tests

func TestScanDir_ReturnsEntriesForValidScripts(t *testing.T) {
	// Scans the whole testdir — invalid-*.sh fixtures will produce warnings,
	// which is expected. We only assert on the entry count here.
	entries, _, err := cmd.ScanDir(testdir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// valid-single.sh → 1 entry, valid-multi.sh → 3 entries
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}
}

func TestScanDir_SkipsNoBlockFiles(t *testing.T) {
	entries, warnings, err := cmd.ScanDir(testdir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// no-block.sh must be silently skipped — no entry, no warning
	for _, e := range entries {
		if filepath.Base(e.Path) == "no-block.sh" {
			t.Error("no-block.sh should be silently skipped")
		}
	}
	for _, w := range warnings {
		if filepath.Base(w) == "no-block.sh" {
			t.Error("no-block.sh should produce no warning")
		}
	}
}

func TestScanDir_WarnsOnInvalidScripts(t *testing.T) {
	entries, warnings, err := cmd.ScanDir(testdir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// invalid-*.sh fixtures have blocks that fail validation
	if len(warnings) == 0 {
		t.Error("expected at least one warning for invalid fixtures")
	}
	// invalid scripts must not produce index entries
	for _, e := range entries {
		base := filepath.Base(e.Path)
		if len(base) > 8 && base[:8] == "invalid-" {
			t.Errorf("invalid script %q should not produce an index entry", base)
		}
	}
}

func TestScanDir_NonExistentDir(t *testing.T) {
	_, _, err := cmd.ScanDir("/nonexistent/path", false)
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}