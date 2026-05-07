package pipeline_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sidisinsane/hashfm-agent/internal/pipeline"
)

const testdir = "../../testdata"

func readFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(testdir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}

func TestProcess_SingleCommand(t *testing.T) {
	block, err := pipeline.Process(readFile(t, "valid-single.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block.IsMulti() {
		t.Fatal("expected single-command block")
	}
	cmd := block.Single
	if cmd.Description == "" {
		t.Error("description is empty")
	}
	if cmd.Usage == "" {
		t.Error("usage is empty")
	}
	if len(cmd.Exits) == 0 {
		t.Error("exits is empty")
	}
	if _, ok := cmd.Exits["0"]; !ok {
		t.Error("expected exit code 0")
	}
}

func TestProcess_MultiCommand(t *testing.T) {
	block, err := pipeline.Process(readFile(t, "valid-multi.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !block.IsMulti() {
		t.Fatal("expected multi-command block")
	}
	if len(block.Multi) != 3 {
		t.Errorf("expected 3 subcommands, got %d", len(block.Multi))
	}
	for i, cmd := range block.Multi {
		if cmd.Description == "" {
			t.Errorf("entry %d: description is empty", i)
		}
		if cmd.Usage == "" {
			t.Errorf("entry %d: usage is empty", i)
		}
		if len(cmd.Exits) == 0 {
			t.Errorf("entry %d: exits is empty", i)
		}
	}
}

func TestProcess_NoBlock(t *testing.T) {
	_, err := pipeline.Process(readFile(t, "no-block.sh"))
	if err == nil {
		t.Fatal("expected ErrNoBlock, got nil")
	}
	if _, ok := err.(pipeline.ErrNoBlock); !ok {
		t.Errorf("expected ErrNoBlock, got %T: %v", err, err)
	}
}

func TestProcess_MissingDescription(t *testing.T) {
	_, err := pipeline.Process(readFile(t, "invalid-missing-description.sh"))
	if err == nil {
		t.Fatal("expected validation error for missing description")
	}
	if !strings.Contains(err.Error(), "description") {
		t.Errorf("error should mention 'description', got: %v", err)
	}
}

func TestProcess_MissingUsage(t *testing.T) {
	_, err := pipeline.Process(readFile(t, "invalid-missing-usage.sh"))
	if err == nil {
		t.Fatal("expected validation error for missing usage")
	}
	if !strings.Contains(err.Error(), "usage") {
		t.Errorf("error should mention 'usage', got: %v", err)
	}
}

func TestProcess_MissingExits(t *testing.T) {
	_, err := pipeline.Process(readFile(t, "invalid-missing-exits.sh"))
	if err == nil {
		t.Fatal("expected validation error for missing exits")
	}
	if !strings.Contains(err.Error(), "exits") {
		t.Errorf("error should mention 'exits', got: %v", err)
	}
}

func TestProcess_MultiTooFewEntries(t *testing.T) {
	_, err := pipeline.Process(readFile(t, "invalid-multi-single-entry.sh"))
	if err == nil {
		t.Fatal("expected validation error for single-entry multi block")
	}
}