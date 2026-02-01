#!/bin/bash

# Setup Kubernetes access to Azure VM K3s cluster
# This script:
# 1. Creates SSH tunnel for K8s API (port 6443)
# 2. Exports KUBECONFIG to use the tunnel
# 3. Port-forwards API server service to localhost:8080

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBECONFIG_PATH="$SCRIPT_DIR/k3s-config/k3s-local.yaml"
VM_HOST="kloudlite@104.211.92.93"
K8S_PORT=6443
API_SERVER_PORT=8080

echo "🔧 Setting up Kubernetes access..."
echo ""

# Kill existing port forwards and tunnels
echo "🧹 Cleaning up existing connections..."
pkill -f "ssh.*6443.*${VM_HOST}" || true
pkill -f "kubectl port-forward.*api-server.*${API_SERVER_PORT}" || true
sleep 2

# Create SSH tunnel for K8s API
echo "🔌 Creating SSH tunnel for K8s API (localhost:${K8S_PORT})..."
ssh -f -N -L ${K8S_PORT}:localhost:${K8S_PORT} ${VM_HOST}
sleep 1

# Export KUBECONFIG
echo "📝 Setting KUBECONFIG..."
export KUBECONFIG="$KUBECONFIG_PATH"

# Verify K8s connection
echo "✅ Verifying K8s connection..."
if ! kubectl get nodes &>/dev/null; then
    echo "❌ Failed to connect to K8s cluster"
    exit 1
fi

echo "✅ K8s cluster connected"
kubectl get nodes

# Port-forward API server
echo ""
echo "🔌 Port-forwarding API server (localhost:${API_SERVER_PORT})..."
kubectl port-forward -n kloudlite svc/api-server ${API_SERVER_PORT}:80 &
PORT_FORWARD_PID=$!
sleep 2

# Verify API server
echo "✅ Verifying API server connection..."
if ! curl -s http://localhost:${API_SERVER_PORT}/healthz &>/dev/null; then
    echo "⚠️  API server not responding yet (this is normal on first start)"
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "🎯 Services available:"
echo "  - K8s API: https://localhost:${K8S_PORT}"
echo "  - API Server: http://localhost:${API_SERVER_PORT}"
echo ""
echo "📝 To use kubectl in this session, run:"
echo "  export KUBECONFIG=\"$KUBECONFIG_PATH\""
echo ""
echo "🛑 To stop the port-forwards:"
echo "  pkill -f 'ssh.*6443'"
echo "  pkill -f 'kubectl port-forward.*api-server'"
echo ""
echo "🔄 Port-forward PID: $PORT_FORWARD_PID"
