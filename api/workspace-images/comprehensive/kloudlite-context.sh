#!/bin/sh
# kloudlite-context.sh - Display Kloudlite workspace context for Starship

# Exit if not in a Kloudlite workspace
[ -z "$WORKSPACE_NAME" ] && exit 0
[ -z "$WORKSPACE_NAMESPACE" ] && exit 0

# Use kubectl with in-cluster config
KUBECTL="kubectl"

# Get connected environment
ENV_NAME=""
if ENV_INFO=$(${KUBECTL} get workspace "$WORKSPACE_NAME" -n "$WORKSPACE_NAMESPACE" -o jsonpath='{.status.connectedEnvironment.name}' 2>/dev/null); then
    if [ -n "$ENV_INFO" ]; then
        ENV_NAME="$ENV_INFO"
    fi
fi

# Get intercepted services
INTERCEPTS=""
if INTERCEPT_LIST=$(${KUBECTL} get serviceintercept -n "$WORKSPACE_NAMESPACE" -l "workspaces.kloudlite.io/workspace=$WORKSPACE_NAME" -o jsonpath='{range .items[*]}{.spec.serviceName}{","}{end}' 2>/dev/null); then
    if [ -n "$INTERCEPT_LIST" ]; then
        # Remove trailing comma
        INTERCEPTS=$(echo "$INTERCEPT_LIST" | sed 's/,$//')
    fi
fi

# Format output
OUTPUT=""

if [ -n "$ENV_NAME" ]; then
    OUTPUT="env:$ENV_NAME"
fi

if [ -n "$INTERCEPTS" ]; then
    # Convert comma-separated to space-separated
    INTERCEPT_NAMES=$(echo "$INTERCEPTS" | tr ',' ' ')
    if [ -n "$OUTPUT" ]; then
        OUTPUT="$OUTPUT | intercepts:$INTERCEPT_NAMES"
    else
        OUTPUT="intercepts:$INTERCEPT_NAMES"
    fi
fi

[ -n "$OUTPUT" ] && echo "$OUTPUT"
