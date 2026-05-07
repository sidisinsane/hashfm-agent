# hashfm-agent

Extracts hashfms from shell scripts and produces a machine-readable index for
LLM agents.

`hashfm-agent` is the primary implementation of the `hashfm` convention. It
reads `# ---` delimited YAML blocks from scripts and generates an index — a
minimal, token-efficient map of available tools.

---

## Install

```bash
# macOS (Homebrew)
brew install sidisinsane/tap/hashfm-agent

# Or download a binary from the releases page
```

---

## Usage

**Generate an index from the current directory:**

```bash
hashfm-agent generate
```

**Set persistent defaults in `.hashfm`:**

```yaml
version: "1.0"
project:
  name: "my-project"

hashfm-agent:
  generate:
    format: tsv
    recursive: true
```

See `.hashfm` in this repo for a working example.

**CLI flags override config file defaults:**

```bash
hashfm-agent generate --format jsonl --output index.jsonl
```

---

## Output Formats

| Format | Token cost | Use case |
|--------|-----------|---------|
| `tsv` (default) | lowest | LLM agent consumption |
| `jsonl` | medium | Pipelines, further processing |
| `yaml` | highest | Human readability |

---

## The `hashfm` Convention

`hashfm-agent` reads a small YAML payload from scripts:

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

A hashfm – hash + frontmatter. See [`hashfm`](https://github.com/sidisinsane/hashfm) for the full specification.

---

## Documentation

| Document | What it covers |
|----------|---------------|
| `spec.md` | hashfm syntax, fields, index formats, agent workflow |
| [`hashfm/CONFIG.md`](https://github.com/sidisinsane/hashfm/blob/main/CONFIG.md) | `.hashfm` config file design and implementation |

---

## Development

### Prerequisites

- Go 1.26+
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
