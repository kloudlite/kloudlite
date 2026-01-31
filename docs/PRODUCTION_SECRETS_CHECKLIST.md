# Production Secrets Checklist

Use this checklist to set up all required secrets in the GitHub production environment.

## Setup Instructions

1. Go to: `https://github.com/kloudlite/kloudlite/settings/environments`
2. Click **New environment** (or select existing `production` environment)
3. Name: `production`
4. Add each secret listed below

## Required Secrets for Production Environment

### ✅ Azure Deployment

- [ ] `AZURE_CREDENTIALS` - Azure service principal for deployment
  ```json
  {
    "clientId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "clientSecret": "your-client-secret",
    "subscriptionId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "tenantId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  }
  ```

### ✅ Application URLs

- [ ] `NEXT_PUBLIC_BASE_URL` - Base URL for console app (e.g., `https://console.kloudlite.io`)
- [ ] `NEXT_PUBLIC_API_URL` - API endpoint URL (e.g., `https://api.kloudlite.io`)

### ✅ Authentication

- [ ] `AUTH_SECRET` - NextAuth secret (generate with: `openssl rand -base64 32`)

### ✅ GitHub OAuth

- [ ] `GITHUB_CLIENT_ID` - GitHub OAuth App Client ID
- [ ] `GITHUB_CLIENT_SECRET` - GitHub OAuth App Client Secret

**Setup GitHub OAuth App**:
1. Go to: `https://github.com/settings/developers`
2. New OAuth App
3. Authorization callback URL: `https://console.kloudlite.io/api/auth/callback/github`

### ✅ Google OAuth

- [ ] `GOOGLE_CLIENT_ID` - Google OAuth Client ID
- [ ] `GOOGLE_CLIENT_SECRET` - Google OAuth Client Secret

**Setup Google OAuth**:
1. Go to: `https://console.cloud.google.com/apis/credentials`
2. Create OAuth 2.0 Client ID
3. Authorized redirect URI: `https://console.kloudlite.io/api/auth/callback/google`

### ✅ Microsoft Entra (Azure AD) OAuth

- [ ] `MICROSOFT_ENTRA_CLIENT_ID` - Microsoft Entra Application (client) ID
- [ ] `MICROSOFT_ENTRA_CLIENT_SECRET` - Microsoft Entra Client Secret
- [ ] `MICROSOFT_ENTRA_TENANT_ID` - Microsoft Entra Tenant ID

**Setup Microsoft Entra**:
1. Go to: `https://portal.azure.com/#blade/Microsoft_AAD_RegisteredApps/ApplicationsListBlade`
2. New registration
3. Redirect URI: `https://console.kloudlite.io/api/auth/callback/azure-ad`

### ✅ Supabase

- [ ] `SUPABASE_URL` - Supabase project URL (e.g., `https://xxxxx.supabase.co`)
- [ ] `SUPABASE_ANON_KEY` - Supabase anonymous/public key
- [ ] `SUPABASE_SERVICE_ROLE_KEY` - Supabase service role key (admin access)

**Get Supabase Credentials**:
1. Go to your Supabase project: `https://app.supabase.com/project/<project-id>/settings/api`
2. Copy URL and keys from Project Settings → API

### ✅ Cloudflare

- [ ] `CLOUDFLARE_API_TOKEN` - Cloudflare API token with DNS edit permissions
- [ ] `CLOUDFLARE_ZONE_ID` - Cloudflare Zone ID for your domain
- [ ] `CLOUDFLARE_DNS_DOMAIN` - Base domain for installations (e.g., `khost.dev`)
- [ ] `CLOUDFLARE_ORIGIN_CA_KEY` - Cloudflare Origin CA API key

**Get Cloudflare Credentials**:
1. **API Token**: Go to `https://dash.cloudflare.com/profile/api-tokens`
   - Create Token → Edit zone DNS template
   - Zone: DNS Edit permissions
2. **Zone ID**: Go to your domain overview, copy Zone ID from right sidebar
3. **Origin CA Key**: Go to `https://dash.cloudflare.com/profile/api-tokens` → Origin CA Key

## Optional: Environment Protection Rules

Configure deployment protection in GitHub:

1. In production environment settings, enable:
   - [ ] **Required reviewers** - Add team members who should approve deployments
   - [ ] **Wait timer** - Optional delay before deployment (e.g., 5 minutes)
   - [ ] **Deployment branches** - Select "Selected branches" → Add `development` and `master`

## Verification

After adding all secrets, verify deployment works:

1. Push a change to `development` branch
2. Check GitHub Actions for the workflow run
3. Verify deployment job waits for environment approval (if enabled)
4. Approve the deployment (if required reviewers are set)
5. Check Azure Container App updated successfully:
   ```bash
   az containerapp show \
     --name ca-kloudlite-registration \
     --resource-group rg-kloudlite-registration \
     --query "properties.template.containers[0].env" -o table
   ```

## Secret Rotation Schedule

| Secret | Rotation Frequency | Last Rotated |
|--------|-------------------|--------------|
| OAuth Secrets | Quarterly | - |
| `AUTH_SECRET` | When team members leave | - |
| API Tokens | Quarterly | - |
| Service Keys | Annually | - |

## Troubleshooting

### "Secret not found" error

- ✅ Ensure secrets are added to the **production environment**, not repository secrets
- ✅ Check secret names match exactly (case-sensitive)
- ✅ Verify workflow has `environment: production` in the deployment job

### OAuth authentication not working

- ✅ Verify callback URLs match in OAuth provider settings
- ✅ Check `NEXT_PUBLIC_BASE_URL` matches actual domain
- ✅ Ensure OAuth client IDs and secrets are correct

### Deployment fails

- ✅ Check `AZURE_CREDENTIALS` is valid and has permissions
- ✅ Verify Azure resource names are correct
- ✅ Check Azure Container App exists and is accessible

## Security Reminders

- 🔒 Never commit secrets to the repository
- 🔒 Rotate secrets regularly
- 🔒 Limit access to environment secrets (admins only)
- 🔒 Enable GitHub secret scanning
- 🔒 Use separate secrets for development and production
- 🔒 Document secret rotation in the table above

---

**Need Help?**
- GitHub Environments: https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment
- See full documentation: [GITHUB_ENVIRONMENT_SETUP.md](./GITHUB_ENVIRONMENT_SETUP.md)
