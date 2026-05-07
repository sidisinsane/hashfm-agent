package generator_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sidisinsane/hashfm-agent/internal/generator"
	"github.com/sidisinsane/hashfm-agent/internal/index"
	"github.com/sidisinsane/hashfm-agent/internal/model"
	"github.com/sidisinsane/hashfm-agent/internal/pipeline"
)

const testdir = "../../testdata"
const tsvHeader = "name\tpath\tdescription"

func testPath(name string) string {
	return filepath.Join(testdir, name)
}

func loadEntries(t *testing.T) []model.IndexEntry {
	t.Helper()
	fixtures := []string{"valid-single.sh", "valid-multi.sh"}
	var entries []model.IndexEntry
	for _, name := range fixtures {
		path := testPath(name)
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		block, err := pipeline.Process(string(src))
		if err != nil {
			t.Fatalf("process %s: %v", name, err)
		}
		entries = append(entries, index.FromBlock(block, path)...)
	}
	return entries
}

func TestTSV(t *testing.T) {
	entries := loadEntries(t)
	var buf bytes.Buffer
	err := generator.TSV{}.Generate(&buf, entries)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != len(entries)+1 {
		t.Fatalf("expected %d lines, got %d", len(entries)+1, len(lines))
	}
	if lines[0] != tsvHeader {
		t.Errorf("wrong header: got %q, want %q", lines[0], tsvHeader)
	}
	for i, line := range lines[1:] {
		if strings.Count(line, "\t") != 2 {
			t.Errorf("line %d: expected 2 tabs, got %d: %q", i+1, strings.Count(line, "\t"), line)
		}
	}
}

func TestTSV_Empty(t *testing.T) {
	var buf bytes.Buffer
	generator.TSV{}.Generate(&buf, nil)
	if !strings.HasPrefix(buf.String(), "name\t") {
		t.Error("expected header even for empty entries")
	}
}

func TestJSONL(t *testing.T) {
	entries := loadEntries(t)
	var buf bytes.Buffer
	err := generator.JSONL{}.Generate(&buf, entries)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != len(entries) {
		t.Fatalf("expected %d lines, got %d", len(entries), len(lines))
	}
	for i, line := range lines {
		var obj map[string]string
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("line %d: invalid JSON: %v", i, err)
			continue
		}
		for _, field := range []string{"name", "path", "description"} {
			if obj[field] == "" {
				t.Errorf("line %d: missing field %q", i, field)
			}
		}
	}
}

func TestYAML(t *testing.T) {
	entries := loadEntries(t)
	var buf bytes.Buffer
	err := generator.YAML{}.Generate(&buf, entries)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "- name:") {
		t.Error("expected YAML list entries starting with '- name:'")
	}
	// valid-multi.sh produces subcommand names containing spaces — must be quoted
	for _, e := range entries {
		if strings.Contains(e.Name, " ") && !strings.Contains(out, `"`+e.Name+`"`) {
			t.Errorf("name with spaces %q should be quoted in YAML output", e.Name)
		}
	}
	// valid-single.sh produces a plain name with no spaces — must not be quoted
	for _, e := range entries {
		if !strings.Contains(e.Name, " ") && strings.Contains(out, `"`+e.Name+`"`) {
			t.Errorf("name without spaces %q should not be quoted in YAML output", e.Name)
		}
	}
}