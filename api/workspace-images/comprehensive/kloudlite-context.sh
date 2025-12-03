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

# Read file content
CONTENT=$(cat "$CONTEXT_FILE" 2>/dev/null)

# Extract environment name using grep and sed
ENV_NAME=$(echo "$CONTENT" | grep -o '"environment":"[^"]*"' | sed 's/"environment":"//;s/"$//')

# Format output
OUTPUT=""

if [ -n "$ENV_NAME" ]; then
    OUTPUT="env:$ENV_NAME"
fi

# Parse intercepts with port mappings using simple grep/sed
# JSON format: {"intercepts":[{"serviceName":"web","portMappings":[{"servicePort":80,"workspacePort":8080}]}]}
# Output format: [web:80->:8080]

# Check if intercepts exist
if echo "$CONTENT" | grep -qF 'serviceName'; then
    # Extract service name
    SVC_NAME=$(echo "$CONTENT" | grep -o '"serviceName":"[^"]*"' | head -1 | sed 's/"serviceName":"//;s/"$//')

    # Extract servicePort and workspacePort
    SVC_PORT=$(echo "$CONTENT" | grep -o '"servicePort":[0-9]*' | head -1 | sed 's/"servicePort"://')
    WS_PORT=$(echo "$CONTENT" | grep -o '"workspacePort":[0-9]*' | head -1 | sed 's/"workspacePort"://')

    if [ -n "$SVC_NAME" ] && [ -n "$SVC_PORT" ] && [ -n "$WS_PORT" ]; then
        INTERCEPTS="[$SVC_NAME:$SVC_PORT->:$WS_PORT]"

        if [ -n "$OUTPUT" ]; then
            OUTPUT="$OUTPUT $INTERCEPTS"
        else
            OUTPUT="$INTERCEPTS"
        fi
    fi
fi

[ -n "$OUTPUT" ] && echo "$OUTPUT"
