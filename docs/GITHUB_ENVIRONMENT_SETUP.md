# GitHub Production Environment Setup

This guide explains how to configure GitHub environment secrets for automated deployments to Azure Container Apps.

## Overview

The deployment workflows use GitHub Environments to manage secrets and environment variables. All sensitive configuration is stored in GitHub and automatically injected into Azure Container Apps during deployment.

## Setting Up the Production Environment

### 1. Create the Production Environment

1. Go to your GitHub repository: `https://github.com/kloudlite/kloudlite/settings/environments`
2. Click **New environment**
3. Name it: `production`
4. Configure protection rules (optional but recommended):
   - ✅ Required reviewers (add team members who should approve production deployments)
   - ✅ Deployment branches: Select "Selected branches" → Add `development` and `master`

### 2. Add Environment Secrets

Navigate to the production environment settings and add the following secrets:

#### Common Secrets

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `AZURE_CREDENTIALS` | Azure service principal credentials for deployment | JSON object with clientId, clientSecret, subscriptionId, tenantId |
| `AUTH_SECRET` | NextAuth.js secret for session encryption | Generate with: `openssl rand -base64 32` |

#### Console Application Secrets

| Secret Name | Description | Required |
|-------------|-------------|----------|
| `NEXT_PUBLIC_BASE_URL` | Base URL for console app | ✅ Yes (e.g., `https://console.kloudlite.io`) |
| `NEXT_PUBLIC_API_URL` | API endpoint URL | ✅ Yes (e.g., `https://api.kloudlite.io`) |
| `GITHUB_CLIENT_ID` | GitHub OAuth App Client ID | ✅ Yes |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth App Client Secret | ✅ Yes |
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID | ✅ Yes |
| `GOOGLE_CLIENT_SECRET` | Google OAuth Client Secret | ✅ Yes |
| `MICROSOFT_ENTRA_CLIENT_ID` | Microsoft Entra (Azure AD) Client ID | ✅ Yes |
| `MICROSOFT_ENTRA_CLIENT_SECRET` | Microsoft Entra Client Secret | ✅ Yes |
| `MICROSOFT_ENTRA_TENANT_ID` | Microsoft Entra Tenant ID | ✅ Yes |
| `SUPABASE_URL` | Supabase project URL | ✅ Yes |
| `SUPABASE_ANON_KEY` | Supabase anonymous key | ✅ Yes |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase service role key (admin access) | ✅ Yes |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token for DNS management | ✅ Yes |
| `CLOUDFLARE_ZONE_ID` | Cloudflare zone ID | ✅ Yes |
| `CLOUDFLARE_DNS_DOMAIN` | Base domain for installations | ✅ Yes (e.g., `khost.dev`) |
| `CLOUDFLARE_ORIGIN_CA_KEY` | Cloudflare Origin CA key for SSL certificates | ✅ Yes |

#### Website Application Secrets

| Secret Name | Description | Required |
|-------------|-------------|----------|
| `NEXT_PUBLIC_BASE_URL` | Base URL for website | Optional (defaults to `https://kloudlite.io`) |

## How It Works

### Deployment Flow

1. **Code Push**: Developer pushes code to `development` branch
2. **Build**: GitHub Actions builds Docker images for console, dashboard, and website
3. **Environment Approval**: If required reviewers are configured, deployment waits for approval
4. **Deployment**:
   - Workflow reads secrets from the production environment
   - Builds a JSON object with all environment variables
   - Calls `scripts/deploy/deploy-to-azure.sh` with the image and environment variables
   - Script updates the Azure Container App with new image and environment variables
5. **Verification**: Container app restarts with updated configuration

### Deployment Script

The deployment script (`scripts/deploy/deploy-to-azure.sh`) handles:
- Image updates
- Environment variable injection
- Deployment verification

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
  '{"NEXT_PUBLIC_API_URL":"https://api.kloudlite.io","AUTH_SECRET":"secret123"}'
```

## Environment Variables Mapping

### Console App

| GitHub Secret | Container App Environment Variable |
|---------------|-----------------------------------|
| `NEXT_PUBLIC_BASE_URL` | `NEXT_PUBLIC_BASE_URL` |
| `NEXT_PUBLIC_API_URL` | `NEXT_PUBLIC_API_URL` |
| `GITHUB_CLIENT_ID` | `GITHUB_CLIENT_ID` |
| `GITHUB_CLIENT_SECRET` | `GITHUB_CLIENT_SECRET` |
| `GOOGLE_CLIENT_ID` | `GOOGLE_CLIENT_ID` |
| `GOOGLE_CLIENT_SECRET` | `GOOGLE_CLIENT_SECRET` |
| `MICROSOFT_ENTRA_CLIENT_ID` | `MICROSOFT_ENTRA_CLIENT_ID` |
| `MICROSOFT_ENTRA_CLIENT_SECRET` | `MICROSOFT_ENTRA_CLIENT_SECRET` |
| `MICROSOFT_ENTRA_TENANT_ID` | `MICROSOFT_ENTRA_TENANT_ID` |
| `AUTH_SECRET` | `AUTH_SECRET` |
| `NEXT_PUBLIC_BASE_URL` | `NEXTAUTH_URL` |
| `SUPABASE_URL` | `NEXT_PUBLIC_SUPABASE_URL` |
| `SUPABASE_ANON_KEY` | `NEXT_PUBLIC_SUPABASE_ANON_KEY` |
| `SUPABASE_SERVICE_ROLE_KEY` | `SUPABASE_SERVICE_ROLE_KEY` |
| `CLOUDFLARE_API_TOKEN` | `CLOUDFLARE_API_TOKEN` |
| `CLOUDFLARE_ZONE_ID` | `CLOUDFLARE_ZONE_ID` |
| `CLOUDFLARE_DNS_DOMAIN` | `CLOUDFLARE_DNS_DOMAIN` |
| `CLOUDFLARE_ORIGIN_CA_KEY` | `CLOUDFLARE_ORIGIN_CA_KEY` |
| `CLOUDFLARE_DNS_DOMAIN` | `NEXT_PUBLIC_INSTALLATION_DOMAIN` |

### Website App

| GitHub Secret | Container App Environment Variable |
|---------------|-----------------------------------|
| `NEXT_PUBLIC_BASE_URL` | `NEXT_PUBLIC_BASE_URL` |
| `AUTH_SECRET` | `AUTH_SECRET` |
| `NEXT_PUBLIC_BASE_URL` | `NEXTAUTH_URL` |

## Security Best Practices

1. **Environment Protection**:
   - Enable required reviewers for production deployments
   - Restrict deployment branches to `development` and `master`
   - Use deployment protection rules to prevent unauthorized deployments

2. **Secret Rotation**:
   - Rotate OAuth secrets regularly (quarterly recommended)
   - Rotate `AUTH_SECRET` when team members leave
   - Update API tokens when permissions change

3. **Access Control**:
   - Limit who can modify environment secrets (repository admins only)
   - Enable audit logging for secret access
   - Use GitHub's secret scanning to detect leaked secrets

4. **Secret Management**:
   - Never commit secrets to the repository
   - Use separate secrets for development and production
   - Document all secrets in this file (description only, not values)

## Troubleshooting

### Deployment fails with "Secret not found"

**Solution**: Ensure all required secrets are added to the production environment, not repository secrets.

### Environment variables not updating in Azure

**Solution**:
1. Check that the deployment workflow has `environment: production` set
2. Verify the secret names match exactly (case-sensitive)
3. Check Azure Container App logs for environment variable injection errors

### OAuth authentication not working

**Solution**:
1. Verify OAuth redirect URIs are configured correctly in provider settings
2. Check that `NEXT_PUBLIC_BASE_URL` matches the actual domain
3. Ensure `AUTH_SECRET` is set and is a valid base64 string

## Manual Deployment

To manually deploy with environment variables:

```bash
# 1. Login to Azure
az login

# 2. Build environment variables JSON
ENV_JSON='{"NEXT_PUBLIC_API_URL":"https://api.kloudlite.io","AUTH_SECRET":"your-secret"}'

# 3. Deploy
./scripts/deploy/deploy-to-azure.sh \
  "ca-kloudlite-registration" \
  "rg-kloudlite-registration" \
  "ghcr.io/kloudlite/kloudlite/web-console:latest" \
  "$ENV_JSON"
```

## Monitoring Deployments

### GitHub Actions

View deployment status:
1. Go to **Actions** tab in GitHub
2. Click on "Build, Push and Deploy Web" workflow
3. Check the deployment jobs (deploy-console, deploy-website)

### Azure Container Apps

Verify deployment:
```bash
# Check container app status
az containerapp show \
  --name ca-kloudlite-registration \
  --resource-group rg-kloudlite-registration \
  --query "properties.runningStatus" -o tsv

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

## Additional Resources

- [GitHub Environments Documentation](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment)
- [Azure Container Apps Documentation](https://learn.microsoft.com/en-us/azure/container-apps/)
- [NextAuth.js Environment Variables](https://next-auth.js.org/configuration/options#environment-variables)
