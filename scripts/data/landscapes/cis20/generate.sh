#!/usr/bin/env bash
set -euo pipefail

# Filters developer-portal-backend/scripts/data/landscapes/cis20/ops2go.json according to:
# - Keep if object name contains 'staging', 'integrate', 'canary', hotfix' (case-insensitive)
# - Keep if 'ccf-universe' value equals one of: 'staging', 'integrate', 'canary', 'hotfix', or 'live' (case-insensitive)
# - Keep if 'update-process' value equals 'live' (case-insensitive)
#
# Outputs the result to 'filtered_ops2go.json' in the same directory.

# Ensure jq is available
if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required but not installed." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_FILE="${SCRIPT_DIR}/ops2go.json"
OUT_FILE="${SCRIPT_DIR}/filtered_ops2go.json"

if [[ ! -f "$SRC_FILE" ]]; then
  echo "Error: ops2go.json not found at $SRC_FILE. Create this file, paste the ops2go.json content here, and re-run the script." >&2
  exit 1
fi


jq '
  with_entries(
    select(
      (.key | test("staging|integrate|canary|hotfix"; "i"))
      or
      ((.value["ccf-universe"]? // "" | ascii_downcase) | IN("staging", "integrate", "canary", "hotfix", "live"))
      or
      ((.value["update-process"]? // "" | ascii_downcase) == "live")
    )
  )
' "$SRC_FILE" > "$OUT_FILE"

echo "Filtered JSON written to: $OUT_FILE"

# Generate landscapes.yaml from filtered_ops2go.json
YAML_OUT="${SCRIPT_DIR}/landscapes.yaml"
{
  echo "landscapes:"
  jq -r '
    to_entries
    | sort_by(.key)
    | .[] as $e
    | ($e.value["ccf-universe"]? // "" | ascii_downcase) as $u
    | ($e.value["update-process"]? // "" | ascii_downcase) as $up
    | ($e.key | ascii_downcase) as $k
    | (
        if ($u == "staging" or ($k|test("staging"))) then "staging"
        elif ($u == "integrate" or ($k|test("integrate"))) then "integrate"
        elif ($u == "canary" or ($k|test("canary"))) then "canary"
        elif ($u == "hotfix" or ($k|test("hotfix"))) then "hotfix"
        elif ($u == "live" or $up == "live") then "live"
        else null end
      ) as $env
    | if ($env == null) then
        halt_error(1)
      else
        "  - name: \($e.key|@json)"
        , "    title: \($e.value.displayname_short // "" | @json)"
        , "    description: \($e.value.displayname_full // "" | @json)"
        , "    domain: \($e.value.domain // "" | @json)"
        , "    project: \"cis20\""
        , "    environment: \($env | @json)"
        , "    metadata:"
        , (["IaaS-console","apm-infra-environment","avs-aggregated-monitor","cam-profile-devod","ccf-universe","cockpit","jumpbox","jumpbox2","landscape-repository","region","slack-channel","type"]
            | map("      \(.): " + ( $e.value[.] // "" | @json))
            | .[])
      end
  ' "$OUT_FILE"
} > "$YAML_OUT"

echo "Landscapes YAML written to: $YAML_OUT"
