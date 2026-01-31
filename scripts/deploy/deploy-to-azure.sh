#!/bin/bash

set -e

# Usage: ./deploy-to-azure.sh <app-name> <resource-group> <image> <env-vars-json>
# Example: ./deploy-to-azure.sh ca-kloudlite-registration rg-kloudlite-registration ghcr.io/kloudlite/kloudlite/web-console:sha-abc123 '{"KEY":"value"}'

APP_NAME="$1"
RESOURCE_GROUP="$2"
IMAGE="$3"
ENV_VARS_JSON="$4"

if [ -z "$APP_NAME" ] || [ -z "$RESOURCE_GROUP" ] || [ -z "$IMAGE" ]; then
  echo "Error: Missing required arguments"
  echo "Usage: $0 <app-name> <resource-group> <image> [env-vars-json]"
  exit 1
fi

echo "==================================================="
echo "Azure Container App Deployment"
echo "==================================================="
echo "App Name:        $APP_NAME"
echo "Resource Group:  $RESOURCE_GROUP"
echo "Image:           $IMAGE"
echo "==================================================="

# Build environment variable arguments
ENV_ARGS=""
if [ -n "$ENV_VARS_JSON" ] && [ "$ENV_VARS_JSON" != "{}" ]; then
  echo "Processing environment variables..."

  # Convert JSON to space-separated KEY=VALUE pairs
  while IFS="=" read -r key value; do
    if [ -n "$key" ] && [ -n "$value" ]; then
      # Remove quotes from value if present
      value=$(echo "$value" | sed 's/^"//;s/"$//')
      ENV_ARGS="$ENV_ARGS --set-env-vars ${key}=${value}"
      echo "  - Setting: $key"
    fi
  done < <(echo "$ENV_VARS_JSON" | jq -r 'to_entries[] | "\(.key)=\(.value)"')
fi

# Update the container app
echo ""
echo "Updating container app..."
UPDATE_CMD="az containerapp update \
  --name $APP_NAME \
  --resource-group $RESOURCE_GROUP \
  --image $IMAGE"

# Add environment variables if any
if [ -n "$ENV_ARGS" ]; then
  UPDATE_CMD="$UPDATE_CMD $ENV_ARGS"
fi

# Execute the update
eval $UPDATE_CMD --query "properties.provisioningState" -o tsv

echo ""
echo "==================================================="
echo "Deployment completed successfully!"
echo "==================================================="
