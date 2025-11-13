#!/usr/bin/env bash
set -euo pipefail

# Convert landscapes_org.yaml to landscapes.yaml for specified subfolders.
# For each item under .landscapes:
# - name -> name
# - display_name -> title and description
# - metadata.annotations["landscape/domain"] -> domain
# - project -> the folder name (e.g., "ca" or "usrv")
# - landscape_type -> environment
#
# Requirements:
# - yq (Mike Farah) must be installed: https://mikefarah.gitbook.io/yq/
#   Example install on macOS: brew install yq

if ! command -v yq >/dev/null 2>&1; then
  echo "Error: yq is required but not installed. Install yq (https://mikefarah.gitbook.io/yq/) and re-run." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

convert_dir() {
  local dir="$1"
  local src="${SCRIPT_DIR}/${dir}/landscapes_org.yaml"
  local dst="${SCRIPT_DIR}/${dir}/landscapes.yaml"

  if [[ ! -f "$src" ]]; then
    echo "Skip: source file not found: $src" >&2
    return 0
  fi

  # Perform the mapping and write the result
  yq -o=yaml -I 2 '
    .landscapes // []
    | map(
        (.metadata.links // []) as $links
        | {
            "name": .name,
            "title": (.display_name // ""),
            "description": (.display_name // ""),
            "domain": (.metadata.annotations."landscape/domain" // ""),
            "project": "'"${dir}"'",
            "environment": (.landscape_type // "")
          }
          + {
            "metadata": ({
              "landscape-repository": (.github_config_url // ""),
              "cam-profile-devod": (.cam_profile_url // ""),
              "auditlog": (($links[] | select(.title == "Auditlog") | .url) // ""),
              "prometheus": (($links[] | select(.title == "Prometheus") | .url) // ""),
              "IaaS-console": (($links[] | select(.title == "AWS") | .url) // ""),
              "gardener": (($links[] | select(.title == "Gardener") | .url) // ""),
              "grafana": (($links[] | select(.title == "Grafana") | .url) // ""),
              "kibana": ((($links[] | select(.title == "Kibana") | .url) // ($links[] | select(.title == "Kibana k8s") | .url)) // ""),
              "dynatrace": (($links[] | select(.title == "Dynatrace") | .url) // "")
            } | with_entries(select(.value != "")))
          }
      )
    | {"landscapes": .}
  ' "$src" > "$dst"

  echo "Wrote: $dst"
}

# Convert for both ca and usrv folders
convert_dir "ca"
convert_dir "usrv"
