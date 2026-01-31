#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}GitHub Production Environment Setup${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed${NC}"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Error: Not authenticated with GitHub CLI${NC}"
    echo "Run: gh auth login"
    exit 1
fi

echo -e "${GREEN}✓ GitHub CLI is installed and authenticated${NC}"
echo ""

# Get repository info
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
echo -e "${BLUE}Repository: ${REPO}${NC}"
echo ""

# Create production environment if it doesn't exist
echo -e "${YELLOW}Creating production environment...${NC}"
gh api repos/${REPO}/environments/production -X PUT -f deployment_branch_policy='{"protected_branches":false,"custom_branch_policies":true}' 2>/dev/null || echo "Environment already exists"
echo -e "${GREEN}✓ Production environment ready${NC}"
echo ""

# Function to set a secret
set_secret() {
    local secret_name="$1"
    local secret_description="$2"
    local is_optional="${3:-false}"
    local default_value="${4:-}"

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}${secret_name}${NC}"
    echo -e "${NC}${secret_description}${NC}"

    if [ "$is_optional" = "true" ] && [ -n "$default_value" ]; then
        echo -e "${NC}Default: ${default_value}${NC}"
    fi

    read -p "Enter value (or press Enter to skip): " -s secret_value
    echo ""

    if [ -n "$secret_value" ]; then
        echo "$secret_value" | gh secret set "$secret_name" --env production --repo "$REPO"
        echo -e "${GREEN}✓ Set ${secret_name}${NC}"
    elif [ "$is_optional" = "true" ]; then
        echo -e "${YELLOW}⊘ Skipped ${secret_name}${NC}"
    else
        echo -e "${RED}✗ Warning: ${secret_name} is required but was skipped${NC}"
    fi
    echo ""
}

# Function to set a secret from file (for JSON secrets)
set_secret_from_file() {
    local secret_name="$1"
    local secret_description="$2"
    local example="$3"

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}${secret_name}${NC}"
    echo -e "${NC}${secret_description}${NC}"
    echo -e "${NC}Example format:${NC}"
    echo -e "${NC}${example}${NC}"

    read -p "Enter value (or press Enter to skip): " secret_value

    if [ -n "$secret_value" ]; then
        echo "$secret_value" | gh secret set "$secret_name" --env production --repo "$REPO"
        echo -e "${GREEN}✓ Set ${secret_name}${NC}"
    else
        echo -e "${RED}✗ Warning: ${secret_name} is required but was skipped${NC}"
    fi
    echo ""
}

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Setting up secrets...${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Azure credentials
set_secret_from_file "AZURE_CREDENTIALS" \
    "Azure service principal credentials for deployment" \
    '{"clientId":"xxx","clientSecret":"xxx","subscriptionId":"xxx","tenantId":"xxx"}'

# Application URLs
set_secret "NEXT_PUBLIC_BASE_URL" \
    "Base URL for console app (e.g., https://console.kloudlite.io)" \
    false

set_secret "NEXT_PUBLIC_API_URL" \
    "API endpoint URL (e.g., https://api.kloudlite.io)" \
    false

# Authentication
echo -e "${YELLOW}Generating AUTH_SECRET...${NC}"
AUTH_SECRET=$(openssl rand -base64 32)
echo "$AUTH_SECRET" | gh secret set "AUTH_SECRET" --env production --repo "$REPO"
echo -e "${GREEN}✓ Generated and set AUTH_SECRET${NC}"
echo ""

# GitHub OAuth
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}GitHub OAuth Configuration${NC}"
echo -e "${NC}Setup at: https://github.com/settings/developers${NC}"
echo -e "${NC}Callback URL: https://console.kloudlite.io/api/auth/callback/github${NC}"
echo ""

set_secret "GITHUB_CLIENT_ID" \
    "GitHub OAuth App Client ID" \
    false

set_secret "GITHUB_CLIENT_SECRET" \
    "GitHub OAuth App Client Secret" \
    false

# Google OAuth
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Google OAuth Configuration${NC}"
echo -e "${NC}Setup at: https://console.cloud.google.com/apis/credentials${NC}"
echo -e "${NC}Redirect URI: https://console.kloudlite.io/api/auth/callback/google${NC}"
echo ""

set_secret "GOOGLE_CLIENT_ID" \
    "Google OAuth Client ID" \
    false

set_secret "GOOGLE_CLIENT_SECRET" \
    "Google OAuth Client Secret" \
    false

# Microsoft Entra (Azure AD) OAuth
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Microsoft Entra (Azure AD) OAuth Configuration${NC}"
echo -e "${NC}Setup at: https://portal.azure.com/#blade/Microsoft_AAD_RegisteredApps${NC}"
echo -e "${NC}Redirect URI: https://console.kloudlite.io/api/auth/callback/azure-ad${NC}"
echo ""

set_secret "MICROSOFT_ENTRA_CLIENT_ID" \
    "Microsoft Entra Application (client) ID" \
    false

set_secret "MICROSOFT_ENTRA_CLIENT_SECRET" \
    "Microsoft Entra Client Secret" \
    false

set_secret "MICROSOFT_ENTRA_TENANT_ID" \
    "Microsoft Entra Tenant ID" \
    false

# Supabase
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Supabase Configuration${NC}"
echo -e "${NC}Get credentials from: https://app.supabase.com/project/<project-id>/settings/api${NC}"
echo ""

set_secret "SUPABASE_URL" \
    "Supabase project URL (e.g., https://xxxxx.supabase.co)" \
    false

set_secret "SUPABASE_ANON_KEY" \
    "Supabase anonymous/public key" \
    false

set_secret "SUPABASE_SERVICE_ROLE_KEY" \
    "Supabase service role key (admin access)" \
    false

# Cloudflare
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Cloudflare Configuration${NC}"
echo -e "${NC}API Token: https://dash.cloudflare.com/profile/api-tokens${NC}"
echo -e "${NC}Zone ID: Found in domain overview${NC}"
echo ""

set_secret "CLOUDFLARE_API_TOKEN" \
    "Cloudflare API token with DNS edit permissions" \
    false

set_secret "CLOUDFLARE_ZONE_ID" \
    "Cloudflare Zone ID for your domain" \
    false

set_secret "CLOUDFLARE_DNS_DOMAIN" \
    "Base domain for installations (e.g., khost.dev)" \
    true \
    "khost.dev"

set_secret "CLOUDFLARE_ORIGIN_CA_KEY" \
    "Cloudflare Origin CA API key" \
    false

echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}✓ Setup Complete!${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Review secrets: gh secret list --env production"
echo "2. Configure environment protection rules (optional):"
echo "   - Go to: https://github.com/${REPO}/settings/environments"
echo "   - Click on 'production'"
echo "   - Add required reviewers"
echo "   - Set deployment branches"
echo "3. Test deployment by pushing to development branch"
echo ""
echo -e "${GREEN}Done!${NC}"
