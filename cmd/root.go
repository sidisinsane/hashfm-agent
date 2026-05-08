// Package cmd implements the hashfm-agent CLI.
//
// # Flag ordering
//
// Go's standard [flag] package stops parsing at the first non-flag argument.
// The Bash prototype used POSIX-style interspersed flags (e.g. generate <dir>
// --format jsonl). This implementation follows Go convention instead:
//
//	hashfm-agent generate [flags] <dir>
//
// All flags must appear before the positional argument. This avoids the
// complexity of a custom pre-parser and is idiomatic for Go CLIs built on
// the standard library.
package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sidisinsane/hashfm-agent/internal/config"
	"github.com/sidisinsane/hashfm-agent/internal/pipeline"
)

const usage = `Usage:
  hashfm-agent generate [flags] <dir>
  hashfm-agent parse <file>

Commands:
  generate    Scan a directory and generate an index of all hashfm-agent scripts
  parse       Parse and validate a single script, print the block as JSON

Generate flags:
  -f, --format    Output format: tsv (default), jsonl, yaml
  -o, --output    Write index to file instead of stdout
  -r, --recursive Scan subdirectories (default: false)
  -c, --config    Path to config file (default: .hashfm)

Note: flags must precede the <dir> argument (Go flag convention).
`

// Execute initializes and runs the CLI application by parsing the command line arguments.
func Execute() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		runGenerate(os.Args[2:])
	case "parse":
		runParse(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}

func runGenerate(args []string) {
	fset := flag.NewFlagSet("generate", flag.ExitOnError)

	var configPath string
	fset.StringVar(&configPath, "config", ".hashfm", "Path to config file")
	fset.StringVar(&configPath, "c", ".hashfm", "Path to config file (shorthand)")

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var format string
	defaultFormat := "tsv"
	if cfg.Agent.Generate.Format != "" {
		defaultFormat = cfg.Agent.Generate.Format
	}
	fset.StringVar(&format, "format", defaultFormat, "Output format")
	fset.StringVar(&format, "f", defaultFormat, "Output format (shorthand)")

	var output string
	defaultOutput := ""
	if cfg.Agent.Generate.Output != "" {
		defaultOutput = cfg.Agent.Generate.Output
	}
	fset.StringVar(&output, "output", defaultOutput, "Write index to file instead of stdout")
	fset.StringVar(&output, "o", defaultOutput, "Output file (shorthand)")

	var recursive bool
	defaultRecursive := cfg.Agent.Generate.Recursive
	fset.BoolVar(&recursive, "recursive", defaultRecursive, "Scan subdirectories")
	fset.BoolVar(&recursive, "r", defaultRecursive, "Recursive (shorthand)")

	fset.Parse(args)

	if fset.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "generate: <dir> argument required")
		fmt.Fprintln(os.Stderr, "usage: hashfm-agent generate [flags] <dir>")
		os.Exit(1)
	}
	dir := fset.Arg(0)

	gen, err := NewGenerator(format)
	if err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}

	entries, warnings, err := ScanDir(dir, recursive)
	if err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}

	w := io.Writer(os.Stdout)
	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "generate:", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}

	if err := gen.Generate(w, entries); err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}

	// Completion summary
	if output != "" {
		fmt.Fprintf(os.Stderr, "Generated index: %s (%d commands from %d scripts)\n", output, len(entries), len(entries))
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Skipped: %s\n", filepath.Base(w))
	}
}

func runParse(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "parse: <file> argument required")
		os.Exit(1)
	}
	path := args[0]

	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse:", err)
		os.Exit(1)
	}

	block, err := pipeline.Process(string(src))
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse:", err)
		os.Exit(1)
	}

	var v any
	if block.IsMulti() {
		v = block.Multi
	} else {
		v = block.Single
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}