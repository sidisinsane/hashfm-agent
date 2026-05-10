package cmd_test

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/sidisinsane/hashfm-agent/cmd"
	"github.com/sidisinsane/hashfm-agent/internal/generator"
)

//go:embed testdata/*
var testFixtures embed.FS

const testdataDir = "testdata"

func testPath(name string) string {
	return filepath.Join(testdataDir, name)
}

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
	entries, _, err := cmd.ScanDir(testdataDir, false, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// valid-single.sh → 1 entry, valid-multi.sh → 3 entries
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}
}

func TestScanDir_SkipsNoBlockFiles(t *testing.T) {
	entries, warnings, err := cmd.ScanDir(testdataDir, false, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// no-block.sh must be silently skipped — no entry, no warning
	for _, e := range entries {
		if pathBase(e.Path) == "no-block.sh" {
			t.Error("no-block.sh should be silently skipped")
		}
	}
	for _, w := range warnings {
		if pathBase(w) == "no-block.sh" {
			t.Error("no-block.sh should produce no warning")
		}
	}
}

func TestScanDir_WarnsOnInvalidScripts(t *testing.T) {
	entries, warnings, err := cmd.ScanDir(testdataDir, false, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// invalid-*.sh fixtures have blocks that fail validation
	if len(warnings) == 0 {
		t.Error("expected at least one warning for invalid fixtures")
	}
	// invalid scripts must not produce index entries
	for _, e := range entries {
		base := pathBase(e.Path)
		if len(base) > 8 && base[:8] == "invalid-" {
			t.Errorf("invalid script %q should not produce an index entry", base)
		}
	}
}

func TestScanDir_NonExistentDir(t *testing.T) {
	_, _, err := cmd.ScanDir("/nonexistent/path", false, nil, nil)
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestScanDir_ExcludePattern(t *testing.T) {
	entries, _, err := cmd.ScanDir(testdataDir, false, nil, []string{"invalid-*.sh"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range entries {
		base := pathBase(e.Path)
		if len(base) > 8 && base[:8] == "invalid-" {
			t.Errorf("invalid script %q should be excluded", base)
		}
	}
	// valid scripts still included
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}
}

func TestScanDir_IncludePattern(t *testing.T) {
	entries, _, err := cmd.ScanDir(testdataDir, false, []string{"valid-*.sh"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// only valid-single.sh and valid-multi.sh match
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}
	for _, e := range entries {
		base := pathBase(e.Path)
		if base[:6] != "valid-" {
			t.Errorf("expected only valid-*.sh, got %q", base)
		}
	}
}

func TestScanDir_ExcludeTakesPrecedence(t *testing.T) {
	entries, _, err := cmd.ScanDir(testdataDir, false, []string{"valid-*.sh"}, []string{"valid-multi.sh"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if pathBase(entries[0].Path) != "valid-single.sh" {
		t.Errorf("expected valid-single.sh, got %q", pathBase(entries[0].Path))
	}
}

func TestScanDir_InvalidPattern(t *testing.T) {
	// Invalid exclude pattern is silently ignored; files are processed normally.
	_, warnings, err := cmd.ScanDir(testdataDir, false, nil, []string{"["})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// invalid-*.sh still produce warnings despite the invalid exclude pattern
	if len(warnings) == 0 {
		t.Error("expected warnings for invalid scripts")
	}
}

// pathBase returns the base name of a path.
func pathBase(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
