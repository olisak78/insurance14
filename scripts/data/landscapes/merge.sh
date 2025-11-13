#!/usr/bin/env bash
set -euo pipefail

# Merge landscapes.yaml from subfolders (ca, cis20, usrv) into scripts/data/landscapes.yaml
# Output: developer-portal-backend/scripts/data/landscapes.yaml
# Requirements:
# - yq (Mike Farah) must be installed: https://mikefarah.gitbook.io/yq/
#   Example install on macOS: brew install yq

if ! command -v yq >/dev/null 2>&1; then
  echo "Error: yq is required but not installed. Install yq and re-run." >&2
  exit 1
fi

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="$BASE_DIR"
OUT_FILE="$(cd "$BASE_DIR/.." && pwd)/landscapes.yaml"

FILES=()
for d in ca cis20 usrv; do
  f="$SRC_DIR/$d/landscapes.yaml"
  if [[ -f "$f" ]]; then
    FILES+=("$f")
  else
    echo "Info: skipping missing file: $f" >&2
  fi
done

if [[ ${#FILES[@]} -eq 0 ]]; then
  echo "Error: no input landscapes.yaml files found under $SRC_DIR/{ca,cis20,usrv}" >&2
  exit 1
fi

# Merge documents: concatenate .landscapes arrays from all inputs into a single array
# and wrap with top-level key "landscapes".
# Note: Use eval-all (ea) to extract arrays per file, then combine via inputs.
yq ea -o=json '.landscapes[]' "${FILES[@]}" | jq -s '{landscapes: .}' | yq -P -o=yaml > "$OUT_FILE"

echo "Merged landscapes written to: $OUT_FILE"
