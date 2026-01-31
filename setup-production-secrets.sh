#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

REPO="kloudlite/kloudlite"
ENV="production"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                                ║${NC}"
echo -e "${BLUE}║         GitHub Production Environment Setup                    ║${NC}"
echo -e "${BLUE}║         Repository: ${REPO}                            ║${NC}"
echo -e "${BLUE}║                                                                ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}✗ GitHub CLI (gh) is not installed${NC}"
    echo ""
    echo "Install it with:"
    echo "  macOS:   brew install gh"
    echo "  Ubuntu:  sudo apt install gh"
    echo "  Or visit: https://cli.github.com/"
    exit 1
fi

echo -e "${GREEN}✓ GitHub CLI is installed${NC}"

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}✗ Not authenticated with GitHub CLI${NC}"
    echo ""
    echo "Please authenticate first:"
    echo "  gh auth login"
    exit 1
fi

echo -e "${GREEN}✓ GitHub CLI is authenticated${NC}"
echo ""

# Get current user
CURRENT_USER=$(gh api user -q .login)
echo -e "${CYAN}Authenticated as: ${CURRENT_USER}${NC}"
echo ""

# Create production environment
echo -e "${YELLOW}Creating production environment...${NC}"
gh api repos/${REPO}/environments/${ENV} -X PUT \
  -f deployment_branch_policy='{"protected_branches":false,"custom_branch_policies":true}' \
  2>/dev/null || echo -e "${YELLOW}Environment already exists${NC}"

echo -e "${GREEN}✓ Production environment ready${NC}"
echo ""

# Function to set a secret
set_secret() {
    local name="$1"
    local description="$2"
    local is_optional="${3:-false}"
    local default_value="${4:-}"
    local is_multiline="${5:-false}"

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}${name}${NC}"
    echo -e "${NC}${description}${NC}"

    if [ "$is_optional" = "true" ] && [ -n "$default_value" ]; then
        echo -e "${YELLOW}(Optional - Default: ${default_value})${NC}"
    elif [ "$is_optional" = "true" ]; then
        echo -e "${YELLOW}(Optional)${NC}"
    else
        echo -e "${RED}(Required)${NC}"
    fi

    echo ""

    if [ "$is_multiline" = "true" ]; then
        echo "Enter value (paste JSON, then press Ctrl+D on a new line):"
        secret_value=$(cat)
    else
        read -p "Enter value (or press Enter to skip): " -s secret_value
        echo ""
    fi

    if [ -n "$secret_value" ]; then
        echo "$secret_value" | gh secret set "$name" --env "$ENV" --repo "$REPO"
        echo -e "${GREEN}✓ Set ${name}${NC}"
    elif [ "$is_optional" = "true" ]; then
        echo -e "${YELLOW}⊘ Skipped ${name}${NC}"
    else
        echo -e "${RED}⚠ Warning: ${name} is required but was skipped${NC}"
    fi
    echo ""
}

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Step 1: Azure Credentials                                     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}To get Azure credentials, run:${NC}"
echo -e "${CYAN}az ad sp create-for-rbac --name 'github-actions' --role contributor --scopes /subscriptions/{subscription-id} --sdk-auth${NC}"
echo ""

set_secret "AZURE_CREDENTIALS" \
    "Azure service principal credentials (JSON format)" \
    false \
    "" \
    true

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Step 2: Authentication                                        ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Generating AUTH_SECRET automatically...${NC}"
AUTH_SECRET=$(openssl rand -base64 32)
echo "$AUTH_SECRET" | gh secret set "AUTH_SECRET" --env "$ENV" --repo "$REPO"
echo -e "${GREEN}✓ Generated and set AUTH_SECRET${NC}"
echo ""

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Step 3: Application URLs                                      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

set_secret "NEXT_PUBLIC_BASE_URL" \
    "Base URL for console app (e.g., https://console.kloudlite.io)" \
    true \
    "https://console.kloudlite.io"

set_secret "NEXT_PUBLIC_API_URL" \
    "API endpoint URL (e.g., https://api.kloudlite.io)" \
    true \
    "https://api.kloudlite.io"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Step 4: GitHub OAuth                                          ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Setup at: https://github.com/settings/developers${NC}"
echo -e "${YELLOW}Callback URL: https://console.kloudlite.io/api/auth/callback/github${NC}"
echo ""

set_secret "GITHUB_CLIENT_ID" \
    "GitHub OAuth App Client ID" \
    false

set_secret "GITHUB_CLIENT_SECRET" \
    "GitHub OAuth App Client Secret" \
    false

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Step 5: Google OAuth                                          ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Setup at: https://console.cloud.google.com/apis/credentials${NC}"
echo -e "${YELLOW}Redirect URI: https://console.kloudlite.io/api/auth/callback/google${NC}"
echo ""

set_secret "GOOGLE_CLIENT_ID" \
    "Google OAuth Client ID" \
    false

set_secret "GOOGLE_CLIENT_SECRET" \
    "Google OAuth Client Secret" \
    false

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Step 6: Microsoft Entra (Azure AD) OAuth                      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Setup at: https://portal.azure.com/#blade/Microsoft_AAD_RegisteredApps${NC}"
echo -e "${YELLOW}Redirect URI: https://console.kloudlite.io/api/auth/callback/azure-ad${NC}"
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

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Step 7: Supabase                                              ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Get from: https://app.supabase.com/project/<project-id>/settings/api${NC}"
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

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Step 8: Cloudflare                                            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}API Token: https://dash.cloudflare.com/profile/api-tokens${NC}"
echo -e "${YELLOW}Zone ID: Found in domain overview${NC}"
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

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Setup Complete!                                               ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${GREEN}✓ All secrets have been configured${NC}"
echo ""
echo -e "${CYAN}Verifying secrets...${NC}"
echo ""

# List all secrets
gh secret list --env "$ENV" --repo "$REPO"

echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Verify secrets: gh secret list --env production --repo ${REPO}"
echo "2. Run verification workflow: gh workflow run manage-secrets.yml -f action=verify-secrets"
echo "3. Test deployment: gh workflow run deploy-web-production.yml -f app=console -f image-tag=dev"
echo ""
echo -e "${GREEN}Done!${NC}"
