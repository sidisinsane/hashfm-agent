#!/usr/bin/env bash
# Shared helper sourced by generators.
# Reads IR lines from an array and populates:
#   index_rows — array of tab-delimited "name\tpath\tdescription" entries
#
# Usage: source this file after populating ir_lines=(), name="", path=""
# Then call: extract_index_rows

TAB=$'\t'

extract_index_rows() {
  local mode=""
  local current_subcmd=""
  local current_description=""
  index_rows=()

  for line in "${ir_lines[@]}"; do
    local field="${line%%${TAB}*}"
    local value="${line#*${TAB}}"

    case "$field" in
      MODE)
        mode="$value"
        ;;
      CMD)
        # Flush previous command entry if we have one
        if [[ -n "$current_description" ]]; then
          index_rows+=("${current_subcmd}${TAB}${path}${TAB}${current_description}")
          current_subcmd=""
          current_description=""
        fi
        ;;
      DESCRIPTION)
        current_description="$value"
        ;;
      USAGE)
        if [[ "$mode" == "single" ]]; then
          current_subcmd="$name"
        else
          # Extract subcommand name — second word after script name
          local subcmd
          subcmd="${value#* }"      # strip script name
          subcmd="${subcmd%% *}"    # take first word of remainder
          current_subcmd="${name} ${subcmd}"
        fi
        ;;
    esac
  done

  # Flush final entry
  if [[ -n "$current_description" ]]; then
    index_rows+=("${current_subcmd}${TAB}${path}${TAB}${current_description}")
  fi
}
