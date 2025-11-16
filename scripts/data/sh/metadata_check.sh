#!/bin/bash
set -euo pipefail

# Helper script to test metadata: search by conditions, echo and/or open browser with URL
# Requires yq (brew install yq)
INPUT_FILE="${1:-landscapes.yaml}"

# Iterate over matching entries (project=cis20, environment=?)
#and (.metadata | has("oc-prefix") | not)
#and .environment == "live"
yq -r '.landscapes[]
  | select(
                .project == "cis20"
                and (.metadata | has("extension") | not)
                and (.metadata | has("oc-prefix") | not)
              )
  | [.name, .domain, .metadata."oc-prefix"] | @tsv' "$INPUT_FILE" \
| while IFS=$'\t' read -r NAME DOMAIN OC_PREFIX; do
  [[ -z "${DOMAIN:-}" ]] && continue


  #URL="https://operation-console.operationsconsole.cfapps.${DOMAIN}"
  #URL="https://operator.operationsconsole.cfapps.${DOMAIN}"
  #URL="https://operations-console.operationsconsole.cfapps.${DOMAIN}"
  #URL="https://${OC_PREFIX}.cfapps.${DOMAIN}"
  URL="https://cp-control-client.cfapps.${DOMAIN}"
  echo "$NAME $URL"

  # Open a new Chrome window and record its ID
  WIN_ID="$(osascript -l AppleScript -e '
on run argv
  set theURL to item 1 of argv
  tell application "Google Chrome"
    if it is not running then activate
    set theWindow to make new window
    set URL of active tab of theWindow to theURL
    return id of theWindow
  end tell
end run
' "$URL")"

  # Wait until the window is closed
  while true; do
    EXISTS="$(osascript -l AppleScript -e '
on run argv
  set targetId to (item 1 of argv) as integer
  tell application "Google Chrome"
    set found to false
    repeat with w in windows
      if id of w is targetId then
        set found to true
        exit repeat
      end if
    end repeat
  end tell
  if found then
    return "1"
  else
    return "0"
  end if
end run
' "$WIN_ID")"

    if [[ "$EXISTS" == "0" ]]; then
      break
    fi
    sleep 1
  done
done
