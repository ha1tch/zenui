#!/bin/sh
# Release gate: the checkpoint zip must not contain repository metadata or
# local state. Runs after the archive step, so the zip is the real artefact.
set -e
cd "$(dirname "$0")"
V=$(tr -d '[:space:]' < VERSION)
Z="zenui-v$V-checkpoint.zip"
[ -f "$Z" ] || { echo "archive $Z not found"; exit 1; }
BAD=$(unzip -Z1 "$Z" | grep -E '^(\.git/|\.ed-journal\.json$|\.release-state\.json$)' || true)
if [ -n "$BAD" ]; then
    echo "$Z contains repository metadata or local state:"
    echo "$BAD" | head -5
    exit 1
fi
echo "archive clean: $Z"
