#!/bin/sh
# kloudlite-context.sh - Display Kloudlite workspace context for Starship
# Reads from cached state file for fast performance without API calls

# Exit if not in a Kloudlite workspace
[ -z "$WORKSPACE_NAME" ] && exit 0

# Path to cached context file (written by workspace controller)
CONTEXT_FILE="/tmp/kloudlite-context.json"

# Check if context file exists
if [ ! -f "$CONTEXT_FILE" ]; then
    exit 0
fi

# Read and parse JSON (using awk/grep since jq may not be available)
# Extract environment name
ENV_NAME=$(grep -o '"environment":"[^"]*"' "$CONTEXT_FILE" 2>/dev/null | cut -d'"' -f4)

# Extract intercepts array and convert to space-separated list
INTERCEPTS=$(grep -o '"intercepts":\[[^]]*\]' "$CONTEXT_FILE" 2>/dev/null | sed 's/"intercepts":\[//;s/\]//;s/"//g;s/,/ /g')

# Format output
OUTPUT=""

if [ -n "$ENV_NAME" ]; then
    OUTPUT="env:$ENV_NAME"
fi

if [ -n "$INTERCEPTS" ]; then
    # Format intercepts list: convert spaces to commas and wrap in parentheses
    FORMATTED_INTERCEPTS=$(echo "$INTERCEPTS" | sed 's/ /, /g')
    if [ -n "$OUTPUT" ]; then
        OUTPUT="$OUTPUT | intercepts:($FORMATTED_INTERCEPTS)"
    else
        OUTPUT="intercepts:($FORMATTED_INTERCEPTS)"
    fi
fi

[ -n "$OUTPUT" ] && echo "$OUTPUT"
