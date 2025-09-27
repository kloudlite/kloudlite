# RBAC Configuration for Kloudlite Platform

## Overview

This document describes the Role-Based Access Control (RBAC) setup for the Kloudlite platform CRDs and how to manage permissions for the API server.

## Architecture

Instead of using a ServiceAccount (which requires the API server to run in-cluster), we use certificate-based authentication for the API server. This allows the API to run outside the cluster while maintaining secure access.

## Quick Start

### 1. Initial Setup

```bash
# Apply CRDs to the cluster
task apply-crds

# Set up RBAC for the API server
task setup-rbac

# This will create:
# - A Kubernetes user "kloudlite-api" with certificates
# - ClusterRoles with appropriate permissions
# - A kubeconfig file at ./kubeconfig/kloudlite-api-kubeconfig.yaml
```

### 2. Running the API Server

```bash
# Run with the generated kubeconfig
task run-with-k8s-auth

# Or manually:
export KUBECONFIG=./kubeconfig/kloudlite-api-kubeconfig.yaml
go run cmd/server/main.go
```

### 3. Testing RBAC

```bash
# Test that the user has correct permissions
task test-rbac
```

## Adding New CRDs

When you create a new CRD, follow these steps to set up RBAC:

### 1. Generate RBAC Rules

```bash
# Generate RBAC rules for your new CRD
./scripts/generate-crd-rbac.sh <crd-name> <api-group>

# Example for a "teams" CRD:
./scripts/generate-crd-rbac.sh teams platform.kloudlite.io

# This creates rbac/teams-rbac.yaml with appropriate roles
```

### 2. Update User Permissions

Edit `scripts/create-k8s-user.sh` and add permissions for your new CRD:

```yaml
- apiGroups: ["platform.kloudlite.io"]
  resources: ["teams"]  # Your new CRD plural name
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["platform.kloudlite.io"]
  resources: ["teams/status"]
  verbs: ["get", "patch", "update"]
```

### 3. Regenerate User Certificate

```bash
# Recreate the user with updated permissions
task create-k8s-user
```

## RBAC Roles Structure

### Aggregated Roles

We use Kubernetes aggregation to automatically combine permissions:

- **platform-admin**: Full access to all platform resources
- **platform-viewer**: Read-only access to all platform resources
- **platform-editor**: Create/Update access (no delete) to all platform resources

### Per-CRD Roles

Each CRD has three associated ClusterRoles:

1. **platform-admin-{crd}**: Full CRUD operations
2. **platform-viewer-{crd}**: Read-only access
3. **platform-editor-{crd}**: Create/Update operations

These roles are automatically aggregated using labels:
- `rbac.kloudlite.io/aggregate-to-platform-admin: "true"`
- `rbac.kloudlite.io/aggregate-to-platform-viewer: "true"`
- `rbac.kloudlite.io/aggregate-to-platform-editor: "true"`

## User Authentication

The API server authenticates using X.509 certificates:

1. **Certificate Generation**: Uses OpenSSL to create a private key and CSR
2. **Kubernetes Signing**: The CSR is submitted to Kubernetes for signing
3. **Kubeconfig Creation**: A kubeconfig file is generated with the signed certificate

The certificate includes:
- **CN (Common Name)**: The username (e.g., "kloudlite-api")
- **O (Organization)**: Used for group membership (e.g., "platform:admins")

## Security Best Practices

1. **Certificate Rotation**: Regularly rotate certificates (recommended: every 90 days)
2. **Least Privilege**: Grant only necessary permissions
3. **Audit Logging**: Enable Kubernetes audit logging to track API access
4. **Secure Storage**: Store kubeconfig files securely, never commit to git
5. **Environment-Specific**: Use different users/certificates for dev/staging/prod

## Troubleshooting

### Permission Denied Errors

```bash
# Check what permissions the user has
kubectl auth can-i --list --kubeconfig=./kubeconfig/kloudlite-api-kubeconfig.yaml

# Check specific permission
kubectl auth can-i create users --all-namespaces --kubeconfig=./kubeconfig/kloudlite-api-kubeconfig.yaml
```

### Certificate Issues

```bash
# View certificate details
openssl x509 -in kubeconfig/certs/kloudlite-api.crt -text -noout

# Check certificate expiration
openssl x509 -in kubeconfig/certs/kloudlite-api.crt -noout -enddate
```

### Regenerate User

```bash
# If you need to regenerate the user with updated permissions
task create-k8s-user USER_NAME=kloudlite-api
```

## Advanced Configuration

### Custom User Creation

```bash
# Create a user with custom name and namespace
task create-k8s-user USER_NAME=my-api-user NAMESPACE=my-namespace
```

### Multiple Environments

For different environments, create separate users:

```bash
# Development
./scripts/create-k8s-user.sh kloudlite-api-dev kloudlite-dev

# Staging
./scripts/create-k8s-user.sh kloudlite-api-staging kloudlite-staging

# Production
./scripts/create-k8s-user.sh kloudlite-api-prod kloudlite-prod
```

## Related Files

- `scripts/create-k8s-user.sh`: Creates Kubernetes users with certificates
- `scripts/generate-crd-rbac.sh`: Generates RBAC rules for new CRDs
- `rbac/`: Directory containing all RBAC YAML definitions
- `kubeconfig/`: Directory containing generated kubeconfig files (gitignored)
- `Taskfile.yml`: Task definitions for RBAC management