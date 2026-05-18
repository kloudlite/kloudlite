#!/bin/bash
# Check Go files are formatted

STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')

if [ -z "$STAGED_GO_FILES" ]; then
    exit 0
fi

UNFORMATTED=$(gofmt -l $STAGED_GO_FILES 2>/dev/null)

if [ -n "$UNFORMATTED" ]; then
    echo "Unformatted Go files:"
    echo "$UNFORMATTED"
    echo ""
    echo "Run: gofmt -w <file>"
    exit 1
fi

exit 0
