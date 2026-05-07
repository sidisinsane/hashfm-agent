package index_test

import (
	"path/filepath"
	"testing"

	"github.com/sidisinsane/hashfm-agent/internal/index"
	"github.com/sidisinsane/hashfm-agent/internal/model"
)

const testdir = "../../testdata"

var exitSuccess = map[string]string{"0": "success"}

func testPath(name string) string {
	return filepath.Join(testdir, name)
}

func TestFromBlock_Single(t *testing.T) {
	path := testPath("valid-single.sh")
	block := model.Block{
		Single: &model.Command{
			Description: "Converts all PNG files in a directory to WebP format",
			Usage:       "convert-images.sh <input_dir> [--quality <0-100>]",
			Exits:       exitSuccess,
		},
	}

	entries := index.FromBlock(block, path)

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Name != "valid-single" {
		t.Errorf("name: got %q, want %q", e.Name, "valid-single")
	}
	if e.Path != path {
		t.Errorf("path: got %q, want %q", e.Path, path)
	}
	if e.Description != block.Single.Description {
		t.Errorf("description: got %q, want %q", e.Description, block.Single.Description)
	}
}

func TestFromBlock_Multi(t *testing.T) {
	path := testPath("valid-multi.sh")
	block := model.Block{
		Multi: []model.Command{
			{Description: "Create and push a feature branch", Usage: "valid-multi.sh feature <branch-name>", Exits: exitSuccess},
			{Description: "Prune merged branches", Usage: "valid-multi.sh cleanup [--dry-run]", Exits: exitSuccess},
			{Description: "Rebase onto main", Usage: "valid-multi.sh sync [--no-push]", Exits: exitSuccess},
		},
	}

	entries := index.FromBlock(block, path)

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	want := []struct{ name, desc string }{
		{"valid-multi feature", "Create and push a feature branch"},
		{"valid-multi cleanup", "Prune merged branches"},
		{"valid-multi sync", "Rebase onto main"},
	}
	for i, w := range want {
		if entries[i].Name != w.name {
			t.Errorf("[%d] name: got %q, want %q", i, entries[i].Name, w.name)
		}
		if entries[i].Description != w.desc {
			t.Errorf("[%d] description: got %q, want %q", i, entries[i].Description, w.desc)
		}
		if entries[i].Path != path {
			t.Errorf("[%d] path: got %q, want %q", i, entries[i].Path, path)
		}
	}
}

func TestFromBlock_ScriptNameStripsExtension(t *testing.T) {
	path := testPath("valid-single.sh")
	block := model.Block{
		Single: &model.Command{
			Description: "Does something",
			Usage:       "valid-single.sh",
			Exits:       exitSuccess,
		},
	}
	entries := index.FromBlock(block, path)
	if entries[0].Name != "valid-single" {
		t.Errorf("got %q, want %q", entries[0].Name, "valid-single")
	}
}