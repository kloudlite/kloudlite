#!/bin/bash

set -e

# Script to restart frontend deployment in k3s cluster after dashboard deployment
# Usage: ./restart-frontend-deployment.sh

RG="${RESOURCE_GROUP:-kl-f615f505-00a6-40b9-b462-400e48c7d224-rg}"
VM_NAME="${VM_NAME:-kl-f615f505-00a6-40b9-b462-400e48c7d224-vm}"
NAMESPACE="${NAMESPACE:-kloudlite}"
DEPLOYMENT="${DEPLOYMENT:-frontend}"

echo "============================================"
echo "Frontend Deployment Restart Script"
echo "============================================"
echo "Resource Group: $RG"
echo "VM Name: $VM_NAME"
echo "Namespace: $NAMESPACE"
echo "Deployment: $DEPLOYMENT"
echo "============================================"
echo ""

# Step 1: Get kubeconfig from VM
echo "📥 Fetching kubeconfig from VM..."
KUBECONFIG_CONTENT=$(az vm run-command invoke \
  --name "$VM_NAME" \
  --resource-group "$RG" \
  --command-id RunShellScript \
  --scripts "sudo cat /etc/rancher/k3s/k3s.yaml" \
  --query 'value[0].message' \
  --output tsv | grep -A 100 "apiVersion: v1")

if [ -z "$KUBECONFIG_CONTENT" ]; then
  echo "❌ Failed to fetch kubeconfig from VM"
  exit 1
fi

echo "✓ Kubeconfig fetched successfully"

# Step 2: Get VM public IP
echo "📡 Getting VM public IP..."
VM_IP=$(az vm show \
  --name "$VM_NAME" \
  --resource-group "$RG" \
  --show-details \
  --query publicIps \
  --output tsv)

if [ -z "$VM_IP" ]; then
  echo "❌ Failed to get VM IP"
  exit 1
fi

echo "✓ VM IP: $VM_IP"

# Step 3: Create temporary kubeconfig with correct server URL
TEMP_KUBECONFIG=$(mktemp)
trap "rm -f $TEMP_KUBECONFIG" EXIT

echo "$KUBECONFIG_CONTENT" | sed "s|https://127.0.0.1:6443|https://$VM_IP:6443|g" > "$TEMP_KUBECONFIG"

echo "✓ Kubeconfig configured with public IP"

# Step 4: Restart the deployment
echo ""
echo "🔄 Restarting $DEPLOYMENT deployment in $NAMESPACE namespace..."

export KUBECONFIG="$TEMP_KUBECONFIG"

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &>/dev/null; then
  echo "❌ Namespace '$NAMESPACE' not found"
  echo "Available namespaces:"
  kubectl get namespaces
  exit 1
fi

# Check if deployment exists
if ! kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" &>/dev/null; then
  echo "❌ Deployment '$DEPLOYMENT' not found in namespace '$NAMESPACE'"
  echo "Available deployments in $NAMESPACE:"
  kubectl get deployments -n "$NAMESPACE"
  exit 1
fi

# Perform rollout restart
kubectl rollout restart deployment/"$DEPLOYMENT" -n "$NAMESPACE"

# Wait for rollout to complete
echo "⏳ Waiting for rollout to complete..."
kubectl rollout status deployment/"$DEPLOYMENT" -n "$NAMESPACE" --timeout=5m

echo ""
echo "✅ Frontend deployment restarted successfully!"
echo ""
echo "Current deployment status:"
kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE"
echo ""
echo "Recent pods:"
kubectl get pods -n "$NAMESPACE" -l app="$DEPLOYMENT" --sort-by=.metadata.creationTimestamp | tail -5
