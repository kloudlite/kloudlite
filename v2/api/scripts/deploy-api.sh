#!/bin/bash
set -e

# Script to deploy the API server to Kubernetes (works everywhere)

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

API_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$API_DIR"

echo -e "${GREEN}Deploying Kloudlite API to Kubernetes${NC}"

# Step 1: Build the Docker image
echo -e "${YELLOW}Building Docker image...${NC}"
docker build -t kloudlite-api:latest .

# Step 2: Load image into K3s (if using K3s)
if command -v k3s &> /dev/null; then
    echo -e "${YELLOW}Loading image into K3s...${NC}"
    docker save kloudlite-api:latest | docker exec -i kloudlite-k3s ctr images import -
elif command -v kind &> /dev/null && kind get clusters | grep -q "kind"; then
    echo -e "${YELLOW}Loading image into Kind...${NC}"
    kind load docker-image kloudlite-api:latest
elif command -v minikube &> /dev/null && minikube status | grep -q "Running"; then
    echo -e "${YELLOW}Loading image into Minikube...${NC}"
    minikube image load kloudlite-api:latest
else
    echo -e "${YELLOW}Using local Docker registry (ensure imagePullPolicy: Never in deployment)${NC}"
fi

# Step 3: Apply CRDs
echo -e "${YELLOW}Applying CRDs...${NC}"
kubectl apply -f crds/

# Step 4: Deploy the API server
echo -e "${YELLOW}Deploying API server...${NC}"
kubectl apply -f deploy/k8s-deployment.yaml

# Step 5: Wait for deployment to be ready
echo -e "${YELLOW}Waiting for deployment to be ready...${NC}"
kubectl wait --for=condition=available --timeout=60s deployment/kloudlite-api -n kloudlite-system

# Step 6: Deploy webhook configurations (optional for now due to TLS requirement)
echo -e "${YELLOW}Note: Webhook configuration requires TLS setup${NC}"
echo -e "${YELLOW}To enable webhooks, configure TLS and apply: kubectl apply -f deploy/webhook-config.yaml${NC}"

# Step 7: Get service information
echo -e "${GREEN}✅ API Server deployed successfully!${NC}"
echo ""
echo -e "${YELLOW}Service Information:${NC}"
kubectl get svc kloudlite-api -n kloudlite-system

# Step 8: Port-forward for local access (optional)
echo ""
echo -e "${YELLOW}To access the API locally, run:${NC}"
echo "kubectl port-forward -n kloudlite-system svc/kloudlite-api 8080:8080"

# Step 9: Check logs
echo ""
echo -e "${YELLOW}To check logs, run:${NC}"
echo "kubectl logs -n kloudlite-system deployment/kloudlite-api -f"