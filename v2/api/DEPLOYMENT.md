# API Server Deployment Guide

This guide provides standardized deployment instructions that work across different environments.

## 🚀 Quick Start

### Option 1: Deploy to Kubernetes (Recommended)

```bash
# 1. Start K3s cluster
task k3s

# 2. Deploy API to Kubernetes
task deploy

# 3. Access the API
task port-forward
# API will be available at http://localhost:8080
```

### Option 2: Run Locally (Development)

```bash
# Run with local kubeconfig
task run
# API will be available at http://localhost:8080
```

## 📦 Deployment Methods

### 1. In-Cluster Deployment (Production Ready)

The API server runs as a pod inside Kubernetes with proper RBAC and service discovery.

```bash
# Deploy everything
task deploy

# This will:
# - Build Docker image
# - Load image into cluster (K3s/Kind/Minikube)
# - Apply CRDs
# - Deploy API server with RBAC
# - Create Service for internal access
```

**Benefits:**
- ✅ Consistent networking across environments
- ✅ Proper RBAC with ServiceAccount
- ✅ Service discovery works automatically
- ✅ Ready for webhook integration
- ✅ Works on any Kubernetes cluster

### 2. Local Development

For quick development iterations:

```bash
# Run locally with external kubeconfig
task run-with-k8s-auth
```

## 🔧 Configuration

### Environment Variables

The API server is configured via ConfigMap when deployed to Kubernetes:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `ENVIRONMENT` | `development` | Environment (development/production) |
| `LOG_LEVEL` | `debug` | Log level (debug/info/warn/error) |
| `KUBERNETES_IN_CLUSTER` | `true` | Use in-cluster config when deployed |
| `KUBERNETES_DEFAULT_NAMESPACE` | `default` | Default namespace for operations |

### Networking

The standardized setup ensures consistent networking:

- **Inside Kubernetes**: `http://kloudlite-api.kloudlite-system.svc.cluster.local:8080`
- **Local Access**: `kubectl port-forward -n kloudlite-system svc/kloudlite-api 8080:8080`
- **NodePort Access**: Available on ports 30000-30100 (if configured)

## 🪝 Webhook Configuration

### Current Status

Webhooks are implemented and tested but require TLS for Kubernetes integration:

- ✅ Validation webhook: `/webhooks/validate/users`
- ✅ Mutation webhook: `/webhooks/mutate/users`
- ⚠️ TLS required for production use

### Testing Webhooks

```bash
# Test webhook logic locally
task test-webhooks
```

### Enabling Webhooks in Kubernetes

1. Configure TLS certificates
2. Update `deploy/webhook-config.yaml` with CA bundle
3. Apply webhook configuration:
   ```bash
   kubectl apply -f deploy/webhook-config.yaml
   ```

## 🛠️ Common Tasks

### View Logs

```bash
task logs
# or
kubectl logs -n kloudlite-system deployment/kloudlite-api -f
```

### Update Deployment

```bash
# Rebuild and redeploy
task deploy
```

### Test RBAC Permissions

```bash
task test-rbac
```

### Clean Up

```bash
# Delete deployment
kubectl delete -f deploy/k8s-deployment.yaml

# Delete CRDs (caution: deletes all User resources)
kubectl delete -f crds/
```

## 🌍 Cross-Environment Compatibility

This setup works consistently across:

- **K3s**: Auto-detects and loads images via `ctr`
- **Kind**: Auto-detects and uses `kind load`
- **Minikube**: Auto-detects and uses `minikube image load`
- **Docker Desktop**: Uses local Docker registry
- **Production Clusters**: Push image to registry and update deployment

## 📝 File Structure

```
v2/api/
├── deploy/
│   ├── k8s-deployment.yaml  # Main deployment manifest
│   └── webhook-config.yaml  # Webhook configuration (requires TLS)
├── scripts/
│   ├── create-k8s-user.sh   # RBAC setup script
│   └── deploy-api.sh        # Deployment automation
├── test/
│   └── test-webhook.sh      # Webhook testing script
└── Taskfile.yml             # Task automation
```

## 🔐 Security Notes

1. **RBAC**: The API server uses a ServiceAccount with minimal required permissions
2. **Webhooks**: Require TLS in production (failurePolicy: Ignore for development)
3. **Network Policies**: Can be added for additional security
4. **Secrets**: Never commit secrets; use Kubernetes Secrets for sensitive data

## 🐛 Troubleshooting

### API not accessible

```bash
# Check if pod is running
kubectl get pods -n kloudlite-system

# Check service
kubectl get svc -n kloudlite-system

# Check logs
kubectl logs -n kloudlite-system deployment/kloudlite-api
```

### Webhook not working

1. Ensure TLS is configured
2. Check webhook configuration: `kubectl get validatingwebhookconfigurations`
3. Check API server logs for webhook requests

### Image not found

```bash
# Ensure image is built
docker images | grep kloudlite-api

# Reload image into cluster
task deploy
```