package config_test

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/sidisinsane/hashfm-agent/internal/config"
)

//go:embed testdata/*
var testFixtures embed.FS

const testdataDir = "testdata"

func readFixture(name string) ([]byte, error) {
	return testFixtures.ReadFile(filepath.Join(testdataDir, name))
}

func TestLoad_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(origCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Agent.Generate.Format != "" {
		t.Errorf("expected empty format, got %q", cfg.Agent.Generate.Format)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(origCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	data, err := readFixture(".hashfm")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := os.WriteFile(".hashfm", data, 0644); err != nil {
		t.Fatalf("write .hashfm: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Agent.Generate.Format != "tsv" {
		t.Errorf("expected format 'tsv', got %q", cfg.Agent.Generate.Format)
	}
	if !cfg.Agent.Generate.Recursive {
		t.Error("expected recursive to be true")
	}
}

func TestLoad_ValidFullConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(origCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	data, err := readFixture("valid-full.yaml")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := os.WriteFile(".hashfm", data, 0644); err != nil {
		t.Fatalf("write .hashfm: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Agent.Generate.Format != "jsonl" {
		t.Errorf("expected format 'jsonl', got %q", cfg.Agent.Generate.Format)
	}
	if cfg.Agent.Generate.Output != "output.jsonl" {
		t.Errorf("expected output 'output.jsonl', got %q", cfg.Agent.Generate.Output)
	}
	if cfg.Agent.Generate.Recursive {
		t.Error("expected recursive to be false")
	}
	if len(cfg.Agent.Generate.Include) != 1 {
		t.Errorf("expected 1 include pattern, got %d", len(cfg.Agent.Generate.Include))
	}
	if len(cfg.Agent.Generate.Exclude) != 1 {
		t.Errorf("expected 1 exclude pattern, got %d", len(cfg.Agent.Generate.Exclude))
	}
}

func TestLoad_NoAgentNamespace(t *testing.T) {
	tmpDir := t.TempDir()
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(origCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	// Write a config with no hashfm-agent namespace
	data := []byte(`version: "1.0"
project:
  name: no-agent-project
`)
	if err := os.WriteFile(".hashfm", data, 0644); err != nil {
		t.Fatalf("write .hashfm: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Agent.Generate.Format != "" {
		t.Errorf("expected empty format, got %q", cfg.Agent.Generate.Format)
	}
}

func TestLoadWithPath_MissingFile(t *testing.T) {
	cfg, err := config.LoadWithPath("/nonexistent/.hashfm")
	if err != nil {
		t.Fatalf("LoadWithPath: unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Agent.Generate.Format != "" {
		t.Errorf("expected empty format, got %q", cfg.Agent.Generate.Format)
	}
}

func TestLoadWithPath_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	data, err := readFixture("valid-full.yaml")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	configPath := filepath.Join(tmpDir, ".hashfm")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.LoadWithPath(configPath)
	if err != nil {
		t.Fatalf("LoadWithPath: unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Agent.Generate.Format != "jsonl" {
		t.Errorf("expected format 'jsonl', got %q", cfg.Agent.Generate.Format)
	}
	if cfg.Agent.Generate.Output != "output.jsonl" {
		t.Errorf("expected output 'output.jsonl', got %q", cfg.Agent.Generate.Output)
	}
}
