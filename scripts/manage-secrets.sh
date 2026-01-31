#!/bin/bash

# Quick script to manage GitHub environment secrets

set -e

ENVIRONMENT="production"

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check gh CLI
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed"
    echo "Install from: https://cli.github.com/"
    exit 1
fi

usage() {
    echo "Usage: $0 <command> [args]"
    echo ""
    echo "Commands:"
    echo "  list                     List all secrets in production environment"
    echo "  set <name> <value>       Set a secret"
    echo "  delete <name>            Delete a secret"
    echo "  view <name>              View secret metadata (not the value)"
    echo "  sync                     Sync secrets from .env file"
    echo ""
    echo "Examples:"
    echo "  $0 list"
    echo "  $0 set API_KEY 'my-secret-value'"
    echo "  $0 delete OLD_SECRET"
    echo "  $0 view AUTH_SECRET"
    exit 1
}

list_secrets() {
    echo -e "${BLUE}Secrets in production environment:${NC}"
    gh secret list --env "$ENVIRONMENT"
}

set_secret() {
    local name="$1"
    local value="$2"

    if [ -z "$name" ] || [ -z "$value" ]; then
        echo "Error: Both name and value are required"
        usage
    fi

    echo "$value" | gh secret set "$name" --env "$ENVIRONMENT"
    echo -e "${GREEN}✓ Set secret: $name${NC}"
}

delete_secret() {
    local name="$1"

    if [ -z "$name" ]; then
        echo "Error: Secret name is required"
        usage
    fi

    gh secret delete "$name" --env "$ENVIRONMENT"
    echo -e "${GREEN}✓ Deleted secret: $name${NC}"
}

view_secret() {
    local name="$1"

    if [ -z "$name" ]; then
        echo "Error: Secret name is required"
        usage
    fi

    gh api "repos/:owner/:repo/environments/$ENVIRONMENT/secrets/$name" | jq
}

sync_from_env() {
    local env_file=".env.production"

    if [ ! -f "$env_file" ]; then
        echo "Error: $env_file not found"
        echo "Create $env_file with your secrets in KEY=VALUE format"
        exit 1
    fi

    echo -e "${YELLOW}Syncing secrets from $env_file...${NC}"

    while IFS='=' read -r key value; do
        # Skip comments and empty lines
        [[ "$key" =~ ^#.*$ ]] && continue
        [[ -z "$key" ]] && continue

        # Remove quotes from value
        value=$(echo "$value" | sed -e 's/^"//' -e 's/"$//' -e "s/^'//" -e "s/'$//")

        echo "$value" | gh secret set "$key" --env "$ENVIRONMENT"
        echo -e "${GREEN}✓ Synced: $key${NC}"
    done < "$env_file"

    echo -e "${GREEN}✓ Sync complete${NC}"
}

# Main
case "${1:-}" in
    list)
        list_secrets
        ;;
    set)
        set_secret "$2" "$3"
        ;;
    delete)
        delete_secret "$2"
        ;;
    view)
        view_secret "$2"
        ;;
    sync)
        sync_from_env
        ;;
    *)
        usage
        ;;
esac
