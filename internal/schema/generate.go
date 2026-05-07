//go:build ignore

// Generate copies schema files from ../schema/ into the current directory.
// This is required because Go's //go:embed cannot reference parent directories.
// Run with: go generate ./internal/schema
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	schemaDir := getSchemaDir()
	srcDir := filepath.Join(schemaDir, "..", "schema")
	dstDir := schemaDir

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: read %s: %v\n", srcDir, err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		src := filepath.Join(srcDir, entry.Name())
		dst := filepath.Join(dstDir, entry.Name())
		if err := copyFile(src, dst); err != nil {
			fmt.Fprintf(os.Stderr, "generate: copy %s: %v\n", entry.Name(), err)
			os.Exit(1)
		}
		fmt.Printf("generate: synced %s\n", entry.Name())
	}
}

func getSchemaDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
