# Deployment Scripts

This directory contains scripts for deploying applications to Azure Container Apps.

## deploy-to-azure.sh

Deploys a Docker image to Azure Container Apps and updates environment variables.

### Usage

```bash
./deploy-to-azure.sh <app-name> <resource-group> <image> [env-vars-json]
```

### Parameters

- `app-name`: Name of the Azure Container App (e.g., `ca-kloudlite-registration`)
- `resource-group`: Azure resource group name (e.g., `rg-kloudlite-registration`)
- `image`: Full Docker image path with tag (e.g., `ghcr.io/kloudlite/kloudlite/web-console:sha-abc123`)
- `env-vars-json` (optional): JSON object containing environment variables

### Examples

#### Deploy without environment variables

```bash
./deploy-to-azure.sh \
  "ca-kloudlite-website" \
  "rg-kloudlite-registration" \
  "ghcr.io/kloudlite/kloudlite/web-website:sha-abc123"
```

#### Deploy with environment variables

```bash
./deploy-to-azure.sh \
  "ca-kloudlite-registration" \
  "rg-kloudlite-registration" \
  "ghcr.io/kloudlite/kloudlite/web-console:sha-abc123" \
  '{
    "NEXT_PUBLIC_API_URL": "https://api.kloudlite.io",
    "AUTH_SECRET": "my-secret-key",
    "GITHUB_CLIENT_ID": "github-client-id"
  }'
```

#### Using in GitHub Actions

```yaml
- name: Deploy to Azure
  run: |
    ENV_JSON='{"API_URL":"${{ secrets.API_URL }}","AUTH_SECRET":"${{ secrets.AUTH_SECRET }}"}'

    ./scripts/deploy/deploy-to-azure.sh \
      "ca-kloudlite-registration" \
      "rg-kloudlite-registration" \
      "ghcr.io/kloudlite/kloudlite/web-console:latest" \
      "$ENV_JSON"
```

### Environment Variables Format

The environment variables must be provided as a JSON object:

```json
{
  "KEY1": "value1",
  "KEY2": "value2",
  "KEY3": "value3"
}
```

Each key-value pair will be converted to:
```bash
--set-env-vars KEY1=value1 KEY2=value2 KEY3=value3
```

### Requirements

- Azure CLI (`az`) must be installed and authenticated
- `jq` must be installed for JSON parsing
- Appropriate permissions to update the Azure Container App

### How It Works

1. **Validates** input parameters
2. **Parses** environment variables from JSON to Azure CLI format
3. **Updates** the container app with:
   - New Docker image
   - Updated environment variables
4. **Verifies** the deployment completed successfully

### Error Handling

The script will exit with an error if:
- Required parameters are missing
- Azure CLI is not authenticated
- Container app update fails
- Invalid JSON format for environment variables

### Debugging

To see detailed output during deployment:

```bash
set -x  # Enable debug mode
./deploy-to-azure.sh <args>
set +x  # Disable debug mode
```

### Security Notes

- Never commit environment variable values to the repository
- Use GitHub secrets for sensitive data
- Environment variables are passed as command-line arguments (visible in process lists during execution)
- Azure Container Apps stores environment variables securely

### Troubleshooting

#### "Command not found: jq"

Install jq:
```bash
# Ubuntu/Debian
sudo apt-get install jq

# macOS
brew install jq

# Alpine
apk add jq
```

#### "Error: The command requires the extension containerapp"

Install the Azure Container Apps extension:
```bash
az extension add --name containerapp --upgrade
```

#### "Deployment timed out"

Increase the timeout or check Azure portal for deployment status:
```bash
az containerapp show \
  --name <app-name> \
  --resource-group <resource-group> \
  --query "properties.provisioningState"
```

## Future Enhancements

- [ ] Support for secret references (Key Vault integration)
- [ ] Rollback functionality
- [ ] Health check validation before completing deployment
- [ ] Support for scaling configuration updates
- [ ] Deployment history tracking
