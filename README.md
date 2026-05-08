# hashfm-agent

Extracts hashfm blocks from shell scripts and produces a machine-readable index
for LLM agents.

`hashfm-agent` is the primary implementation of the `hashfm` convention. It
reads `# ---` delimited YAML blocks from scripts and generates an index — a
minimal, token-efficient map of available tools.

---

## Getting Started

**Install (macOS and Linux):**

```bash
curl -o- https://raw.githubusercontent.com/sidisinsane/hashfm-agent/main/install.sh | bash
```

**Generate an index:**

```bash
# Current directory
hashfm-agent generate .

# With flags
hashfm-agent generate --format jsonl --output index.jsonl .
```

---

## The Hashfm Convention

`hashfm-agent` reads a small YAML payload from scripts delimited by `# ---`:

```bash
#!/usr/bin/env bash
# ---
# description: Deploy the application to staging
# usage: deploy.sh <environment> [--dry-run]
# exits:
#   0: success
#   1: environment not provided
#   2: deployment failed
# ---
```

See [`hashfm`](https://github.com/sidisinsane/hashfm) for the full
specification — delimiters, line prefix, parser rules, and the config file
schema.

---

## Configuration

Set persistent defaults in `.hashfm`:

```yaml
version: "1.1.0"
project:
  name: "my-project"

hashfm-agent:
  generate:
    format: tsv
    recursive: true
    output: index.tsv
```

See [`hashfm/CONFIG.md`](https://github.com/sidisinsane/hashfm/blob/main/CONFIG.md)
for supported filenames, schema, and validation rules.

---

## Output Formats

| Format | Token cost | Use case |
|--------|-----------|---------|
| `tsv` (default) | lowest | LLM agent consumption |
| `jsonl` | medium | Pipelines, further processing |
| `yaml` | highest | Human readability |

---

## Documentation

| Document | What it covers |
|----------|---------------|
| `spec.md` | Agent field schema, index formats, agent workflow |
| [`hashfm/spec.md`](https://github.com/sidisinsane/hashfm/blob/main/spec.md) | Base hashfm syntax and rules |
| [`hashfm/CONFIG.md`](https://github.com/sidisinsane/hashfm/blob/main/CONFIG.md) | `.hashfm` config file design and implementation |

---

## Development

### Prerequisites

- Go 1.26.2
- [golangci-lint](https://golangci-lint.run) — for Go linting
- [lefthook](https://lefthook.dev) — for pre-commit hooks

### Setup

```bash
# Install lefthook
brew install lefthook

# Enable hooks
lefthook install
```

### Linting

Go files are linted with `golangci-lint`. Run manually:

```bash
golangci-lint run ./...
```

Hooks run automatically on commit via lefthook.
