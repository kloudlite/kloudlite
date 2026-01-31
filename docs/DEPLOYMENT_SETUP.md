# Deployment Setup Guide

This guide explains how to set up automated deployments to Azure Container Apps with GitHub Actions and environment management.

## Quick Start

### Option 1: Automated Setup (Recommended)

```bash
# 1. Install GitHub CLI if not already installed
brew install gh  # macOS
# or: sudo apt install gh  # Ubuntu/Debian

# 2. Authenticate with GitHub
gh auth login

# 3. Run the setup script
./scripts/setup-github-secrets.sh
```

The script will:
- Create the production environment in GitHub
- Prompt you for all required secrets
- Auto-generate `AUTH_SECRET`
- Set up all secrets in GitHub

### Option 2: Sync from .env file

```bash
# 1. Copy the template
cp .env.production.template .env.production

# 2. Fill in your values
nano .env.production

# 3. Sync to GitHub
./scripts/manage-secrets.sh sync
```

### Option 3: Manual Setup

See [PRODUCTION_SECRETS_CHECKLIST.md](./PRODUCTION_SECRETS_CHECKLIST.md) for the web interface approach.

## Managing Secrets

### List all secrets

```bash
./scripts/manage-secrets.sh list
```

### Set a single secret

```bash
./scripts/manage-secrets.sh set SECRET_NAME "secret-value"
```

### Update an existing secret

```bash
# Same as set - it will overwrite
./scripts/manage-secrets.sh set AUTH_SECRET "new-secret-value"
```

### Delete a secret

```bash
./scripts/manage-secrets.sh delete OLD_SECRET
```

### View secret metadata

```bash
./scripts/manage-secrets.sh view AUTH_SECRET
```

## Deployment Workflow

### How it works

1. **Code Push**: Push code to `development` branch
2. **Build**: GitHub Actions builds Docker images
3. **Deployment**:
   - Workflow reads secrets from production environment
   - Builds environment variables JSON
   - Calls deployment script with image and environment variables
   - Azure Container App updates with new image and env vars

### Deployment Script

Located at `scripts/deploy/deploy-to-azure.sh`

**Usage**:
```bash
./scripts/deploy/deploy-to-azure.sh \
  <app-name> \
  <resource-group> \
  <image> \
  <env-vars-json>
```

**Example**:
```bash
./scripts/deploy/deploy-to-azure.sh \
  "ca-kloudlite-registration" \
  "rg-kloudlite-registration" \
  "ghcr.io/kloudlite/kloudlite/web-console:sha-abc123" \
  '{"NEXT_PUBLIC_API_URL":"https://api.kloudlite.io"}'
```

## Required Secrets

### Console Application

| Secret | Description |
|--------|-------------|
| `NEXT_PUBLIC_BASE_URL` | Console base URL (e.g., `https://console.kloudlite.io`) |
| `NEXT_PUBLIC_API_URL` | API endpoint (e.g., `https://api.kloudlite.io`) |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret |
| `MICROSOFT_ENTRA_CLIENT_ID` | Microsoft Entra client ID |
| `MICROSOFT_ENTRA_CLIENT_SECRET` | Microsoft Entra client secret |
| `MICROSOFT_ENTRA_TENANT_ID` | Microsoft Entra tenant ID |
| `AUTH_SECRET` | NextAuth secret (auto-generated) |
| `SUPABASE_URL` | Supabase project URL |
| `SUPABASE_ANON_KEY` | Supabase anonymous key |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase service role key |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token |
| `CLOUDFLARE_ZONE_ID` | Cloudflare zone ID |
| `CLOUDFLARE_DNS_DOMAIN` | Base domain (e.g., `khost.dev`) |
| `CLOUDFLARE_ORIGIN_CA_KEY` | Cloudflare Origin CA key |

### Website Application

| Secret | Description |
|--------|-------------|
| `NEXT_PUBLIC_BASE_URL` | Website base URL (defaults to `https://kloudlite.io`) |
| `AUTH_SECRET` | NextAuth secret (shared with console) |

### Azure Deployment

| Secret | Description |
|--------|-------------|
| `AZURE_CREDENTIALS` | Service principal credentials JSON |

## Environment Protection (Optional)

Add deployment protection rules:

```bash
# Via GitHub web UI:
# 1. Go to Settings → Environments → production
# 2. Add required reviewers
# 3. Set deployment branches to: development, master
```

Or configure via API:

```bash
gh api repos/:owner/:repo/environments/production/deployment-branch-policies \
  -X POST \
  -f name=development \
  -f type=branch
```

## Testing Deployment

### 1. Trigger a deployment

```bash
# Make a change to web code
echo "// test" >> web/apps/console/src/app/page.tsx

# Commit and push
git add web/apps/console/src/app/page.tsx
git commit -m "Test deployment"
git push origin development
```

### 2. Monitor deployment

```bash
# Watch GitHub Actions
gh run watch

# Or view in browser
gh run list --workflow="Build, Push and Deploy Web"
```

### 3. Verify in Azure

```bash
# Check container app status
az containerapp show \
  --name ca-kloudlite-registration \
  --resource-group rg-kloudlite-registration \
  --query "properties.runningStatus"

# View environment variables
az containerapp show \
  --name ca-kloudlite-registration \
  --resource-group rg-kloudlite-registration \
  --query "properties.template.containers[0].env" -o table
```

## Troubleshooting

### "gh: command not found"

Install GitHub CLI:
```bash
# macOS
brew install gh

# Ubuntu/Debian
sudo apt install gh

# Or download from: https://cli.github.com/
```

### "Error: Not authenticated"

```bash
gh auth login
# Follow the prompts
```

### Secrets not updating in Azure

1. Check deployment logs in GitHub Actions
2. Verify secret names match exactly (case-sensitive)
3. Check workflow has `environment: production`
4. View Azure Container App logs:
   ```bash
   az containerapp logs show \
     --name ca-kloudlite-registration \
     --resource-group rg-kloudlite-registration \
     --follow
   ```

### OAuth not working after deployment

1. Verify callback URLs in OAuth provider match
2. Check `NEXT_PUBLIC_BASE_URL` is correct
3. Ensure OAuth secrets are set correctly
4. Test OAuth flow in browser and check console errors

## Manual Deployment

To deploy manually outside of GitHub Actions:

```bash
# 1. Login to Azure
az login

# 2. Build environment variables JSON
ENV_JSON=$(cat <<EOF
{
  "NEXT_PUBLIC_API_URL": "https://api.kloudlite.io",
  "AUTH_SECRET": "your-secret"
}
EOF
)

# 3. Deploy
./scripts/deploy/deploy-to-azure.sh \
  "ca-kloudlite-registration" \
  "rg-kloudlite-registration" \
  "ghcr.io/kloudlite/kloudlite/web-console:latest" \
  "$ENV_JSON"
```

## Security Best Practices

### Secret Rotation

Rotate secrets regularly:

```bash
# Generate new AUTH_SECRET
NEW_SECRET=$(openssl rand -base64 32)
./scripts/manage-secrets.sh set AUTH_SECRET "$NEW_SECRET"

# Rotate OAuth secrets
./scripts/manage-secrets.sh set GITHUB_CLIENT_SECRET "new-secret"
```

### Access Control

- Limit who can modify production environment secrets
- Enable required reviewers for production deployments
- Use GitHub audit log to track secret access
- Enable secret scanning in repository settings

### Best Practices

- ✅ Never commit secrets to git
- ✅ Use separate secrets for dev/staging/production
- ✅ Rotate secrets quarterly
- ✅ Use environment-specific OAuth apps
- ✅ Enable 2FA for all team members
- ✅ Audit secret access regularly

## Additional Scripts

### Setup Script
`scripts/setup-github-secrets.sh` - Interactive setup wizard

### Secrets Management
`scripts/manage-secrets.sh` - CLI for managing secrets

### Deployment Script
`scripts/deploy/deploy-to-azure.sh` - Azure Container Apps deployment

## Resources

- [GitHub Environments](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment)
- [GitHub CLI Manual](https://cli.github.com/manual/)
- [Azure Container Apps](https://learn.microsoft.com/en-us/azure/container-apps/)
- [Full Setup Guide](./GITHUB_ENVIRONMENT_SETUP.md)
- [Secrets Checklist](./PRODUCTION_SECRETS_CHECKLIST.md)
