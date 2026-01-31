# Production Deployment Guide

Professional workflow-based deployment system for Kloudlite web applications to Azure Container Apps.

## Overview

The deployment system consists of three main workflows:

1. **Build and Push** (`build-and-push-web.yml`) - Builds Docker images
2. **Production Deployment** (`deploy-web-production.yml`) - Deploys to Azure
3. **Secret Management** (`manage-secrets.yml`) - Validates and guides secret configuration

All deployment logic lives within GitHub Actions workflows - no external scripts required.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHub Actions Workflows                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Build & Push Web                                           │
│  ├─ Build console, dashboard, website (matrix)              │
│  ├─ Push Docker images to GHCR                              │
│  └─ Trigger deployment workflow                             │
│                                                              │
│  Deploy Web Production                                      │
│  ├─ Prepare deployment (determine apps & tags)              │
│  ├─ Deploy Console  ──┐                                     │
│  ├─ Deploy Website   ├──► Reusable workflow                 │
│  ├─ Deploy Dashboard ┘    (_deploy-azure-containerapp.yml)  │
│  └─ Notify (summary report)                                 │
│                                                              │
│  Manage Secrets                                              │
│  ├─ Verify secret configuration                             │
│  ├─ Provide setup guides                                    │
│  └─ Rotation instructions                                   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │ GitHub Environment    │
                │  "production"         │
                │                       │
                │ Secrets:              │
                │ - OAuth configs       │
                │ - API credentials     │
                │ - Azure credentials   │
                └───────────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │ Azure Container Apps  │
                │                       │
                │ - ca-kloudlite-       │
                │   registration        │
                │ - ca-kloudlite-       │
                │   website             │
                │ - ca-kloudlite-       │
                │   dashboard           │
                └───────────────────────┘
```

## Quick Start

### 1. Setup GitHub Environment

Create the production environment:

```bash
# Using GitHub CLI
gh api repos/:owner/:repo/environments/production -X PUT
```

Or via GitHub UI:
1. Go to: `Settings → Environments → New environment`
2. Name: `production`
3. Add protection rules (optional):
   - ✅ Required reviewers
   - ✅ Deployment branches: `development`, `master`

### 2. Configure Secrets

Use GitHub CLI to set secrets:

```bash
# Generate and set AUTH_SECRET
openssl rand -base64 32 | gh secret set AUTH_SECRET --env production

# Set OAuth secrets
echo 'your-github-client-id' | gh secret set GITHUB_CLIENT_ID --env production
echo 'your-github-secret' | gh secret set GITHUB_CLIENT_SECRET --env production

# Set other required secrets
echo 'value' | gh secret set GOOGLE_CLIENT_ID --env production
echo 'value' | gh secret set GOOGLE_CLIENT_SECRET --env production
echo 'value' | gh secret set MICROSOFT_ENTRA_CLIENT_ID --env production
echo 'value' | gh secret set MICROSOFT_ENTRA_CLIENT_SECRET --env production
echo 'value' | gh secret set MICROSOFT_ENTRA_TENANT_ID --env production
echo 'value' | gh secret set SUPABASE_URL --env production
echo 'value' | gh secret set SUPABASE_ANON_KEY --env production
echo 'value' | gh secret set SUPABASE_SERVICE_ROLE_KEY --env production
echo 'value' | gh secret set CLOUDFLARE_API_TOKEN --env production
echo 'value' | gh secret set CLOUDFLARE_ZONE_ID --env production
echo 'value' | gh secret set CLOUDFLARE_DNS_DOMAIN --env production
echo 'value' | gh secret set CLOUDFLARE_ORIGIN_CA_KEY --env production

# Azure credentials (JSON format)
echo '{"clientId":"...","clientSecret":"...","subscriptionId":"...","tenantId":"..."}' | \
  gh secret set AZURE_CREDENTIALS --env production
```

### 3. Verify Configuration

Run the secret verification workflow:

1. Go to: `Actions → Manage Production Secrets → Run workflow`
2. Select action: `verify-secrets`
3. Click "Run workflow"
4. Check the summary for missing secrets

## Required Secrets

### Console Application

| Secret | Required | Default | Description |
|--------|----------|---------|-------------|
| `NEXT_PUBLIC_BASE_URL` | No | `https://console.kloudlite.io` | Console base URL |
| `NEXT_PUBLIC_API_URL` | No | `https://api.kloudlite.io` | API endpoint URL |
| `AUTH_SECRET` | **Yes** | - | NextAuth session secret |
| `GITHUB_CLIENT_ID` | **Yes** | - | GitHub OAuth client ID |
| `GITHUB_CLIENT_SECRET` | **Yes** | - | GitHub OAuth client secret |
| `GOOGLE_CLIENT_ID` | **Yes** | - | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | **Yes** | - | Google OAuth client secret |
| `MICROSOFT_ENTRA_CLIENT_ID` | **Yes** | - | Microsoft Entra client ID |
| `MICROSOFT_ENTRA_CLIENT_SECRET` | **Yes** | - | Microsoft Entra client secret |
| `MICROSOFT_ENTRA_TENANT_ID` | **Yes** | - | Microsoft Entra tenant ID |
| `SUPABASE_URL` | **Yes** | - | Supabase project URL |
| `SUPABASE_ANON_KEY` | **Yes** | - | Supabase anonymous key |
| `SUPABASE_SERVICE_ROLE_KEY` | **Yes** | - | Supabase service role key |
| `CLOUDFLARE_API_TOKEN` | **Yes** | - | Cloudflare API token |
| `CLOUDFLARE_ZONE_ID` | **Yes** | - | Cloudflare zone ID |
| `CLOUDFLARE_DNS_DOMAIN` | No | `khost.dev` | Base domain |
| `CLOUDFLARE_ORIGIN_CA_KEY` | **Yes** | - | Cloudflare Origin CA key |
| `AZURE_CREDENTIALS` | **Yes** | - | Azure service principal JSON |

### Website Application

| Secret | Required | Default | Description |
|--------|----------|---------|-------------|
| `NEXT_PUBLIC_BASE_URL` | No | `https://kloudlite.io` | Website base URL |
| `AUTH_SECRET` | **Yes** | - | NextAuth session secret (shared) |
| `AZURE_CREDENTIALS` | **Yes** | - | Azure service principal JSON |

## Workflows

### Build and Push Web

**File**: `.github/workflows/build-and-push-web.yml`

**Triggers**:
- Push to `development` or `master` branches
- Changes to `web/**`
- Manual trigger

**What it does**:
1. Builds console, dashboard, and website in parallel (matrix strategy)
2. Pushes Docker images to GHCR with tags:
   - `sha-{short-sha}`
   - `dev` (for development branch)
   - `latest` (for master branch)
3. Runs linting checks

**Does NOT deploy** - purely build and push.

### Deploy Web Production

**File**: `.github/workflows/deploy-web-production.yml`

**Triggers**:
- Manual workflow dispatch (select specific app)
- Automatic after successful build (deploys console & website)

**Manual deployment**:
```bash
# Using GitHub CLI
gh workflow run deploy-web-production.yml \
  -f app=console \
  -f image-tag=sha-abc123

# Or via GitHub UI
# Actions → Deploy Web to Production → Run workflow
```

**What it does**:
1. **Prepare**: Determines which apps to deploy and image tag
2. **Deploy**: Calls reusable workflow for each app
   - Updates Azure Container App with new image
   - Injects environment variables from GitHub secrets
   - Verifies deployment status
3. **Notify**: Generates deployment report with status of all apps

**Features**:
- Selective deployment (choose specific app)
- Environment variable injection
- Deployment verification
- Comprehensive status reporting
- Supports environment protection rules

### Reusable Deployment Workflow

**File**: `.github/workflows/_deploy-azure-containerapp.yml`

**Purpose**: Reusable workflow for deploying to Azure Container Apps

**Parameters**:
- `app-name`: Azure Container App name
- `resource-group`: Azure resource group
- `image`: Full Docker image path with tag
- `environment-variables`: JSON object of env vars
- `public-url`: Public URL for summary

**What it does**:
1. Login to Azure
2. Parse environment variables from JSON
3. Update Container App with image and env vars
4. Verify deployment status
5. Generate deployment summary

### Manage Secrets

**File**: `.github/workflows/manage-secrets.yml`

**Triggers**: Manual workflow dispatch only

**Actions**:
- `verify-secrets`: Validate all required secrets are configured
- `update-oauth`: Display OAuth setup guide
- `update-supabase`: Display Supabase configuration guide
- `update-cloudflare`: Display Cloudflare configuration guide
- `rotate-auth-secret`: Guide for rotating AUTH_SECRET
- `setup-all`: Complete setup instructions

**Usage**:
```bash
# Verify secrets
gh workflow run manage-secrets.yml -f action=verify-secrets

# Get OAuth setup guide
gh workflow run manage-secrets.yml -f action=update-oauth
```

## Deployment Workflows

### Automatic Deployment

When code is pushed to `development`:

```
1. Build & Push workflow runs
   ├─ Builds all apps
   ├─ Pushes images with sha-{commit} tag
   └─ Completes successfully

2. Deploy workflow automatically triggers
   ├─ Deploys console with latest image
   ├─ Deploys website with latest image
   └─ Generates deployment report
```

### Manual Deployment

Deploy specific app with specific tag:

```bash
# Deploy console with specific image tag
gh workflow run deploy-web-production.yml \
  -f app=console \
  -f image-tag=sha-abc123

# Deploy all apps
gh workflow run deploy-web-production.yml \
  -f app=all \
  -f image-tag=latest

# Deploy just website
gh workflow run deploy-web-production.yml \
  -f app=website \
  -f image-tag=dev
```

### Rollback Deployment

To rollback to a previous version:

```bash
# Find previous working commit
git log --oneline -10

# Deploy that version
gh workflow run deploy-web-production.yml \
  -f app=console \
  -f image-tag=sha-{previous-commit}
```

## Monitoring Deployments

### Watch deployment in real-time

```bash
# Start deployment
gh workflow run deploy-web-production.yml -f app=all -f image-tag=dev

# Watch progress
gh run watch

# View latest run
gh run view --web
```

### Check deployment status

```bash
# List recent deployments
gh run list --workflow=deploy-web-production.yml --limit 5

# View specific run
gh run view <run-id>
```

### Verify in Azure

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

# View logs
az containerapp logs show \
  --name ca-kloudlite-registration \
  --resource-group rg-kloudlite-registration \
  --follow
```

## Security Best Practices

### Environment Protection

Enable protection rules for production environment:

1. Go to: `Settings → Environments → production`
2. Configure:
   - ✅ **Required reviewers**: Add team members
   - ✅ **Wait timer**: 5 minutes (optional)
   - ✅ **Deployment branches**: `development`, `master` only

### Secret Management

- ✅ Use GitHub environment secrets (not repository secrets)
- ✅ Rotate secrets quarterly
- ✅ Enable GitHub secret scanning
- ✅ Audit secret access via GitHub audit log
- ✅ Use separate OAuth apps for production

### Access Control

- ✅ Limit who can modify production environment
- ✅ Require 2FA for all team members with deployment access
- ✅ Use branch protection rules on deployment branches
- ✅ Enable required status checks before merging

## Troubleshooting

### Deployment fails: "Secret not found"

**Solution**: Ensure secrets are in the **production environment**, not repository secrets.

```bash
# Check environment secrets
gh secret list --env production

# Set missing secret
echo 'value' | gh secret set SECRET_NAME --env production
```

### Environment variables not updating

**Solution**: Secrets are read during workflow run, not at deployment time.

1. Verify secret is set correctly in production environment
2. Trigger a new deployment (don't rerun failed job)
3. Check Azure Container App logs for errors

### OAuth not working after deployment

**Solution**: Verify callback URLs match in OAuth provider.

1. Check `NEXT_PUBLIC_BASE_URL` is correct
2. Verify OAuth callback URLs:
   - GitHub: `https://console.kloudlite.io/api/auth/callback/github`
   - Google: `https://console.kloudlite.io/api/auth/callback/google`
   - Microsoft: `https://console.kloudlite.io/api/auth/callback/azure-ad`

### Deployment stuck on "Waiting for approval"

**Solution**: Environment protection rules require manual approval.

1. Go to: `Actions → Deploy Web to Production → [Running workflow]`
2. Click "Review deployments"
3. Select "production" and approve

## Advanced Usage

### Deploy from different branches

```bash
# Deploy from feature branch
git checkout feature/new-ui
gh workflow run deploy-web-production.yml -f app=console -f image-tag=dev
```

### Custom environment variables

Modify `.github/workflows/deploy-web-production.yml`:

```yaml
environment-variables: |
  {
    "NEXT_PUBLIC_BASE_URL": "${{ secrets.NEXT_PUBLIC_BASE_URL }}",
    "CUSTOM_VAR": "custom-value"
  }
```

### Multiple environments

Create additional environments (staging, development):

```bash
# Create staging environment
gh api repos/:owner/:repo/environments/staging -X PUT

# Set staging secrets
echo 'value' | gh secret set SECRET_NAME --env staging

# Duplicate deployment workflow for staging
cp .github/workflows/deploy-web-production.yml \
   .github/workflows/deploy-web-staging.yml
```

## Support

For issues or questions:
- Check workflow run logs in GitHub Actions
- Review Azure Container App logs
- Verify secrets configuration with `manage-secrets` workflow

---

**Deployment System Version**: 2.0
**Last Updated**: January 2026
