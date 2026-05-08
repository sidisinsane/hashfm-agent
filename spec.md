# hashfm-agent

A minimal, token-efficient interface between Bash scripts and LLM agents. An
implementation of [hashfm](https://raw.githubusercontent.com/sidisinsane/hashfm/refs/heads/main/spec.md).

See the `hashfm` spec for the base convention — delimiters, line prefix, and
parser rules. This document defines the `hashfm-agent` field schema, the index,
output formats, and the agent workflow.

---

## Design Principles

- **Agent-first.** Every element of this convention exists to help an LLM agent
  decide whether and how to invoke a script. Nothing is included for developer
  convenience alone.
- **Token-efficient.** The least information that is sufficient is the right amount.
- **One source of truth.** The script is the SSOT. All other artifacts are
  derived from it.

---

## Hashfm Block Schema

### Single-command script

Used when the script has one purpose and one invocation form.

```bash
#!/usr/bin/env bash
# ---
# description: Converts all PNG files in a directory to WebP format
# usage: valid-single.sh [--quality] [--dry-run]
# exits:
#   0: success
#   1: no PNG files found in working directory
# ---
```

### Multi-command script

Used when the script exposes multiple subcommands. The block contains a YAML
sequence — one entry per subcommand.

```bash
#!/usr/bin/env bash
# ---
# - description: Create and push a feature branch from current HEAD
#   usage: valid-multi.sh feature
# - description: Prune local branches already merged into main
#   usage: valid-multi.sh cleanup
# ---
```

### Parser Disambiguation Rule

If the first content line inside the block starts with `- `, the block is
treated as multi-command. Otherwise it is treated as single-command. These two
forms are mutually exclusive.

---

## Fields

### `description`

**Mandatory.**

A single line describing what the script or subcommand does. Written for an
agent scanning for relevance. Should answer: *what does this do?*

- One line only.
- No trailing punctuation.
- Written in the imperative: *Converts*, *Syncs*, *Rotates* — not *A script that
  converts*.

### `usage`

**Mandatory.**

A single line showing how to invoke the script or subcommand. Includes the
script name, positional arguments, and optional flags using conventional notation.

| Notation | Meaning |
|---|---|
| `arg` | Required positional argument |
| `[arg]` | Optional positional argument |
| `[--flag value]` | Optional flag with value |
| `[--flag]` | Optional boolean flag |

### `exits`

**Mandatory.**

A map of exit codes to plain-language descriptions. Every exit code the script
can produce must be listed. The description answers: *what does this code mean,
and why might it occur?*

```yaml
exits:
  0: success
  1: config file not found at expected path
  2: API request failed — check network connectivity
```

Exit code keys are integers. Values are single lines.

---

## The Index

The index is the only artifact generated from `hashfm-agent` blocks. It is a
single file listing all discovered commands. Its purpose is discovery — an agent
reads the index to find candidate tools before reading any individual script.

### Index Fields

Derived from the block and the filesystem:

| Field | Source |
|---|---|
| `name` | Script filename without extension. For subcommands: `scriptname subcommand`. |
| `path` | Relative path to the script file. |
| `description` | `description` field from the block or subcommand entry. |

For multi-command scripts, each subcommand produces its own index row.

### Output Formats

Formats are ranked by token efficiency.

#### TSV (default — most token-efficient)

A header line followed by one command per line, tab-separated.

```tsv
name	path	description
valid-single	./testdata/valid-single.sh	Converts all PNG files in a directory to WebP format
valid-multi feature	./testdata/valid-multi.sh	Create and push a feature branch from current HEAD
valid-multi cleanup	./testdata/valid-multi.sh	Prune local branches already merged into main
valid-multi sync	./testdata/valid-multi.sh	Rebase current branch onto main and force-push
```

#### JSONL

One JSON object per line.

```jsonl
{"name":"valid-single","path":"./testdata/valid-single.sh","description":"Converts all PNG files in a directory to WebP format"}
{"name":"valid-multi feature","path":"./testdata/valid-multi.sh","description":"Create and push a feature branch from current HEAD"}
{"name":"valid-multi cleanup","path":"./testdata/valid-multi.sh","description":"Prune local branches already merged into main"}
{"name":"valid-multi sync","path":"./testdata/valid-multi.sh","description":"Rebase current branch onto main and force-push"}
```

#### YAML

One list entry per command, no wrapper key.

```yaml
- name: valid-single
  path: ./testdata/valid-single.sh
  description: Converts all PNG files in a directory to WebP format
- name: "valid-multi feature"
  path: ./testdata/valid-multi.sh
  description: Create and push a feature branch from current HEAD
- name: "valid-multi cleanup"
  path: ./testdata/valid-multi.sh
  description: Prune local branches already merged into main
- name: "valid-multi sync"
  path: ./testdata/valid-multi.sh
  description: Rebase current branch onto main and force-push
```

### Future: Tags And Filtered Indexes

A `tags` field is reserved for future use. When implemented it will be an
optional list field in the block. The index generator will support filtering by
tag and producing per-tag sub-indexes. Tags are not part of the current schema
and are ignored by the parser if present.

---

## Agent Workflow

```text
1. DISCOVER   Agent reads the index.
              Scans descriptions to find candidate tools.
              Cost: one file read, minimal tokens.

2. DECIDE     Agent reads the script's hashfm block.
              Checks usage, exits. Constructs the invocation.
              Cost: one file read per candidate.

3. INVOKE     Agent runs the script.

4. RECOVER    Agent reads exit code and stderr.
              Consults exits map to interpret failure.
              Decides next step.
```

The block is read directly from the script. No intermediate extraction document
is produced. The index is the only generated artifact.

---

## Summary Of Fields

| Field | Level | Cardinality | Required |
|---|---|---|---|
| `description` | command | single line | yes |
| `usage` | command | single line | yes |
| `exits` | command | map | yes |
| `tags` | file | list | no (reserved) |

---

## Schema Sync And Config Loading

Both schemas exist in two places:

| Schema | `schema/` (source of truth) | `internal/schema/` (embedded) |
|---|---|---|
| Block — `hashfm-agent.schema.json` | Published on GitHub | `//go:embed` in the binary |
| Config — `hashfm-agent-config.schema.json` | Published on GitHub | `//go:embed` in the binary |

Go's embed directive cannot reference parent directories or URLs. The
`internal/schema/` files are copies, never edited directly. They are synced
automatically via a Go generate directive:

```bash
go generate ./internal/schema
```

This is also run automatically by GoReleaser before each release build. To
manually trigger the sync:

```bash
cd hashfm-agent
go generate ./internal/schema
```

To edit a schema, edit the file in `schema/` and run `go generate` to sync.

### Config Loading

The config file is loaded via `hashfm.LoadConfig`, which provides:

- Config file discovery (`.hashfm`, `.hashfm.yml`, `.hashfm.yaml`, `.hashfm.json`)
- Validation against the core `hashfm-config.schema.json`

The `hashfm-agent` namespace is then validated against
`hashfm-agent-config.schema.json`.

See [`hashfm/CONFIG.md`](https://github.com/sidisinsane/hashfm/blob/main/CONFIG.md)
for the full config file specification.
