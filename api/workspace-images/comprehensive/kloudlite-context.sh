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

# Format output
OUTPUT=""

if [ -n "$ENV_NAME" ]; then
    OUTPUT="env:$ENV_NAME"
fi

# Parse intercepts with port mappings
# JSON format: {"intercepts":[{"serviceName":"web","portMappings":[{"servicePort":80,"workspacePort":8080}]}]}
# Output format: [web:80->:8080, svc2:3948->:8439]

# Check if intercepts array is not empty
if grep -q '"intercepts":\[\{' "$CONTEXT_FILE" 2>/dev/null; then
    # Use awk to parse the intercepts and format them
    INTERCEPTS=$(awk '
    BEGIN { RS=""; FS="" }
    {
        # Extract intercepts array content
        match($0, /"intercepts":\[([^\]]*)\]/, arr)
        if (arr[1] != "") {
            content = arr[1]
            result = ""

            # Split by serviceName to get each intercept
            n = split(content, parts, "\"serviceName\":")
            for (i = 2; i <= n; i++) {
                # Extract service name
                match(parts[i], /^"([^"]+)"/, svc)
                if (svc[1] != "") {
                    svcName = svc[1]

                    # Extract port mappings for this service
                    # Find portMappings array
                    if (match(parts[i], /"portMappings":\[([^\]]*)\]/, pm)) {
                        pmContent = pm[1]
                        # Extract each port mapping
                        numPorts = split(pmContent, portParts, "servicePort\":")
                        for (j = 2; j <= numPorts; j++) {
                            match(portParts[j], /^([0-9]+)/, sp)
                            match(portParts[j], /workspacePort\":([0-9]+)/, wp)
                            if (sp[1] != "" && wp[1] != "") {
                                if (result != "") result = result ", "
                                result = result svcName ":" sp[1] "->:" wp[1]
                            }
                        }
                    }
                }
            }

            if (result != "") print "[" result "]"
        }
    }' "$CONTEXT_FILE")

    if [ -n "$INTERCEPTS" ]; then
        if [ -n "$OUTPUT" ]; then
            OUTPUT="$OUTPUT $INTERCEPTS"
        else
            OUTPUT="$INTERCEPTS"
        fi
    fi
fi

[ -n "$OUTPUT" ] && echo "$OUTPUT"
