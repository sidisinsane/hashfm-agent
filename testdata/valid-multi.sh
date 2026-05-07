#!/usr/bin/env bash
# ---
# - description: Create and push a feature branch from current HEAD
#   usage: git-tool.sh feature <branch-name>
#   exits:
#     0: success
#     1: branch-name not provided
#     2: branch already exists locally or on remote
#     3: not inside a git repository
# - description: Prune local branches already merged into main
#   usage: git-tool.sh cleanup [--dry-run]
#   exits:
#     0: success
#     1: not inside a git repository
#     2: main branch not found
# - description: Rebase current branch onto main and force-push
#   usage: git-tool.sh sync [--no-push]
#   exits:
#     0: success
#     1: not inside a git repository
#     2: rebase conflict detected — resolve manually and re-run
#     3: force-push rejected by remote
# ---

# =============================================================================
# git-tool.sh — Git workflow helpers
#
# Internal implementation below. The tool block above is the agent-facing
# interface. Comments here are for developers.
# =============================================================================

set -euo pipefail

SCRIPT_NAME="$(basename "$0")"

# -----------------------------------------------------------------------------
# Helpers
# -----------------------------------------------------------------------------

die() {
  echo "${SCRIPT_NAME}: error: $*" >&2
  exit 1
}

require_git_repo() {
  git rev-parse --git-dir > /dev/null 2>&1 || die "not inside a git repository"
}

require_arg() {
  local name="$1"
  local value="$2"
  [[ -n "$value" ]] || die "${name} not provided"
}

branch_exists() {
  local branch="$1"
  git show-ref --verify --quiet "refs/heads/${branch}" 2>/dev/null || \
  git ls-remote --exit-code --heads origin "${branch}" > /dev/null 2>&1
}

# -----------------------------------------------------------------------------
# Subcommands
# -----------------------------------------------------------------------------

cmd_feature() {
  local branch_name="${1:-}"
  require_arg "branch-name" "$branch_name"
  require_git_repo

  if branch_exists "$branch_name"; then
    echo "${SCRIPT_NAME}: branch '${branch_name}' already exists locally or on remote" >&2
    exit 2
  fi

  git checkout -b "$branch_name"
  git push -u origin "$branch_name"
  echo "Created and pushed branch: ${branch_name}"
}

cmd_cleanup() {
  local dry_run=false
  [[ "${1:-}" == "--dry-run" ]] && dry_run=true

  require_git_repo

  git show-ref --verify --quiet refs/heads/main 2>/dev/null || {
    echo "${SCRIPT_NAME}: main branch not found" >&2
    exit 2
  }

  local merged
  merged="$(git branch --merged main | grep -v '^\*' | grep -v 'main' || true)"

  if [[ -z "$merged" ]]; then
    echo "No merged branches to prune."
    return 0
  fi

  if "$dry_run"; then
    echo "Would prune:"
    echo "$merged"
  else
    echo "$merged" | xargs git branch -d
    echo "Pruned merged branches."
  fi
}

cmd_sync() {
  local no_push=false
  [[ "${1:-}" == "--no-push" ]] && no_push=true

  require_git_repo

  local current
  current="$(git rev-parse --abbrev-ref HEAD)"

  git fetch origin main

  if ! git rebase origin/main; then
    echo "${SCRIPT_NAME}: rebase conflict detected — resolve manually and re-run" >&2
    exit 2
  fi

  if ! "$no_push"; then
    git push --force-with-lease origin "${current}" || {
      echo "${SCRIPT_NAME}: force-push rejected by remote" >&2
      exit 3
    }
  fi

  echo "Synced ${current} onto main."
}

# -----------------------------------------------------------------------------
# Dispatch
# -----------------------------------------------------------------------------

main() {
  local subcommand="${1:-}"

  case "$subcommand" in
    feature)  shift; cmd_feature "$@" ;;
    cleanup)  shift; cmd_cleanup "$@" ;;
    sync)     shift; cmd_sync "$@" ;;
    *)
      echo "usage: ${SCRIPT_NAME} <subcommand> [options]" >&2
      echo "subcommands: feature, cleanup, sync" >&2
      exit 1
      ;;
  esac
}

main "$@"
