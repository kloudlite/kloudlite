# Deployment Scripts

Scripts for managing GitHub secrets and deploying to Azure Container Apps.

## Quick Start

### Setup GitHub Production Environment

```bash
./setup-github-secrets.sh
```

Interactive wizard that:
- Creates production environment in GitHub
- Prompts for all required secrets
- Auto-generates AUTH_SECRET
- Sets up everything automatically

### Manage Secrets

```bash
# List all secrets
./manage-secrets.sh list

# Set a secret
./manage-secrets.sh set SECRET_NAME "value"

# Delete a secret
./manage-secrets.sh delete SECRET_NAME

# Sync from .env.production file
./manage-secrets.sh sync
```

### Deploy to Azure

```bash
./deploy/deploy-to-azure.sh \
  <app-name> \
  <resource-group> \
  <image> \
  <env-vars-json>
```

## Files

- `setup-github-secrets.sh` - Interactive setup wizard for GitHub environment
- `manage-secrets.sh` - CLI tool for managing individual secrets
- `deploy/deploy-to-azure.sh` - Deploy Docker image to Azure Container Apps
- `deploy/README.md` - Deployment script documentation

## Documentation

- [Deployment Setup Guide](../docs/DEPLOYMENT_SETUP.md) - Complete setup guide
- [GitHub Environment Setup](../docs/GITHUB_ENVIRONMENT_SETUP.md) - Detailed environment configuration
- [Production Secrets Checklist](../docs/PRODUCTION_SECRETS_CHECKLIST.md) - Secret requirements

## Prerequisites

### GitHub CLI

```bash
# macOS
brew install gh

# Ubuntu/Debian
sudo apt install gh

# Authenticate
gh auth login
```

### Azure CLI

```bash
# macOS
brew install azure-cli

# Ubuntu/Debian
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# Authenticate
az login
```

### jq (for JSON parsing)

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt install jq
```

## Examples

### First-time setup

```bash
# 1. Install dependencies
brew install gh azure-cli jq  # macOS

# 2. Authenticate
gh auth login
az login

# 3. Run setup wizard
./setup-github-secrets.sh

# 4. Verify secrets
./manage-secrets.sh list
```

### Update a single secret

```bash
# Update API URL
./manage-secrets.sh set NEXT_PUBLIC_API_URL "https://api.newdomain.com"
```

### Sync from file

```bash
# 1. Copy template
cp ../.env.production.template ../.env.production

# 2. Edit values
nano ../.env.production

# 3. Sync to GitHub
./manage-secrets.sh sync
```

### Manual deployment

```bash
# Build environment variables
ENV_JSON='{"API_URL":"https://api.kloudlite.io","AUTH_SECRET":"secret"}'

# Deploy
./deploy/deploy-to-azure.sh \
  "ca-kloudlite-registration" \
  "rg-kloudlite-registration" \
  "ghcr.io/kloudlite/kloudlite/web-console:latest" \
  "$ENV_JSON"
```

## Troubleshooting

### Command not found: gh

Install GitHub CLI: https://cli.github.com/

### Not authenticated with GitHub

```bash
gh auth login
```

### Not authenticated with Azure

```bash
az login
```

### Permission denied

```bash
chmod +x *.sh
chmod +x deploy/*.sh
```

## Security

- Never commit `.env.production` to git (already in .gitignore)
- Rotate secrets regularly
- Use separate secrets for dev/staging/production
- Enable required reviewers for production deployments
- Audit secret access via GitHub audit log

## Support

For detailed documentation and troubleshooting, see:
- [docs/DEPLOYMENT_SETUP.md](../docs/DEPLOYMENT_SETUP.md)
