#!/usr/bin/env bash
# ---
# description: Converts all PNG files in a directory to WebP format
# usage: convert-images.sh <input_dir> [--quality <0-100>] [--dry-run]
# exits:
#   0: success
#   1: input_dir not provided
#   2: input_dir does not exist or is not a directory
#   3: no PNG files found in input_dir
#   4: cwebp not found — install via package manager
#   5: one or more conversions failed — partial output may exist
# ---

set -euo pipefail

SCRIPT_NAME="$(basename "$0")"
DEFAULT_QUALITY=85

die() {
  echo "${SCRIPT_NAME}: error: $*" >&2
  exit 1
}

# -----------------------------------------------------------------------------
# Parse arguments
# -----------------------------------------------------------------------------

input_dir="${1:-}"
quality="$DEFAULT_QUALITY"
dry_run=false

[[ -n "$input_dir" ]] || die "input_dir not provided"
shift

while [[ $# -gt 0 ]]; do
  case "$1" in
    --quality)
      shift
      quality="${1:-}"
      [[ "$quality" =~ ^[0-9]+$ ]] && [[ "$quality" -le 100 ]] || \
        die "--quality must be an integer between 0 and 100"
      ;;
    --dry-run)
      dry_run=true
      ;;
    *)
      die "unknown option: $1"
      ;;
  esac
  shift
done

# -----------------------------------------------------------------------------
# Validate
# -----------------------------------------------------------------------------

[[ -d "$input_dir" ]] || die "input_dir does not exist or is not a directory"

command -v cwebp > /dev/null 2>&1 || {
  echo "${SCRIPT_NAME}: cwebp not found — install via package manager" >&2
  exit 4
}

mapfile -t pngs < <(find "$input_dir" -maxdepth 1 -name "*.png" | sort)

[[ ${#pngs[@]} -gt 0 ]] || {
  echo "${SCRIPT_NAME}: no PNG files found in ${input_dir}" >&2
  exit 3
}

# -----------------------------------------------------------------------------
# Convert
# -----------------------------------------------------------------------------

failed=0

for png in "${pngs[@]}"; do
  webp="${png%.png}.webp"

  if "$dry_run"; then
    echo "would convert: ${png} → ${webp} (quality: ${quality})"
    continue
  fi

  if cwebp -q "$quality" "$png" -o "$webp" 2>/dev/null; then
    echo "converted: $(basename "$png") → $(basename "$webp")"
  else
    echo "${SCRIPT_NAME}: failed to convert: ${png}" >&2
    (( failed++ )) || true
  fi
done

if [[ "$failed" -gt 0 ]]; then
  echo "${SCRIPT_NAME}: ${failed} conversion(s) failed" >&2
  exit 5
fi
