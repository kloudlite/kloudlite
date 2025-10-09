# WorkMachine Node Manager Deployment Status

## ✅ Completed Components

### 1. **Code Implementation**
- ✅ PackageRequest CRD types (`api/pkg/apis/packages/v1/`)
- ✅ WorkMachine Node Manager application (`api/cmd/workmachine-node-manager/main.go`)
- ✅ Dockerfile for workmachine-node-manager (`api/Dockerfile.workmachine-node-manager`)
- ✅ WorkMachine controller updated to deploy DaemonSet (`api/internal/controllers/workmachine_controller.go:103-109`)
- ✅ Workspace controller updated to create PackageRequests (`api/internal/controllers/workspace_controller.go`)
- ✅ PackageRequest CRD registered in API scheme (`api/internal/controllers/manager.go:37`)

### 2. **Kubernetes Resources**
- ✅ RBAC resources deployed (ServiceAccount, ClusterRole, ClusterRoleBinding)
- ✅ PackageRequest CRD deployed to cluster
- ✅ CRD manifests generated (`devenv/manifests/packages.kloudlite.io_packagerequests.yaml`)
- ✅ RBAC manifest created (`devenv/manifests/workmachine-node-manager-rbac.yaml`)

## 📋 Remaining Steps

### ✅ 1. Build and Push WorkMachine Node Manager Image (COMPLETED)

The Docker image has been built and loaded into k3s:
```bash
cd /Users/karthik/dev/kl-workspace/kloudlite-v2/api
docker build -t kloudlite/workmachine-node-manager:latest -f Dockerfile.workmachine-node-manager .
docker save kloudlite/workmachine-node-manager:latest | docker exec -i k3s-dev ctr images import -
```

**Note**: For the dockerized k3s setup, the image is loaded using `docker exec -i k3s-dev ctr images import -` instead of `sudo k3s ctr images import -`.

### 2. Create Sample Packages File

Create `.kloudlite/packages.yaml` in your workspace repository:
```yaml
packages:
  - name: jq
  - name: curl
  - name: git
  - name: nodejs
    version: "18"  # Optional version specification
```

### 3. Test the Flow

#### Create a WorkMachine:
```bash
kubectl --kubeconfig=devenv/k3s-config/k3s.yaml apply -f - <<EOF
apiVersion: machines.kloudlite.io/v1
kind: WorkMachine
metadata:
  name: test-machine
spec:
  targetNamespace: test-workspace-ns
  desiredState: running
EOF
```

#### Verify DaemonSet is created:
```bash
kubectl --kubeconfig=devenv/k3s-config/k3s.yaml get daemonset package-cache -n default
```

#### Create a Workspace with packages:
```bash
kubectl --kubeconfig=devenv/k3s-config/k3s.yaml apply -f - <<EOF
apiVersion: workspaces.kloudlite.io/v1
kind: Workspace
metadata:
  name: test-workspace
  namespace: test-workspace-ns
spec:
  workspacePath: /workspace
  hostPath: /path/to/your/repo
  packagesFile: .kloudlite/packages.yaml
  workMachineRef:
    name: test-machine
EOF
```

#### Monitor PackageRequest:
```bash
# Watch PackageRequest creation and status
kubectl --kubeconfig=devenv/k3s-config/k3s.yaml get packagerequest -n test-workspace-ns -w

# Check package installation status
kubectl --kubeconfig=devenv/k3s-config/k3s.yaml describe packagerequest -n test-workspace-ns
```

#### Verify packages are installed:
```bash
# Check workspace pod is created after packages are ready
kubectl --kubeconfig=devenv/k3s-config/k3s.yaml get pods -n test-workspace-ns

# Exec into workspace pod and verify packages
kubectl --kubeconfig=devenv/k3s-config/k3s.yaml exec -it <workspace-pod> -n test-workspace-ns -- which jq
```

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         WorkMachine                          │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ State: Running                                         │ │
│  │ TargetNamespace: wm-user                               │ │
│  └────────────────────────────────────────────────────────┘ │
│                             │                                │
│                             ▼                                │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  WorkMachine Node Manager Deployment (targetNamespace)│ │
│  │ ┌─────────────────────────────────────────────────┐   │ │
│  │ │  workmachine-node-manager container              │   │ │
│  │ │  - Watches PackageRequest CRs (namespace-scoped) │   │ │
│  │ │  - Installs packages using Nix                   │   │ │
│  │ │  - Mounts hostPath: /var/lib/kloudlite/nix-store │   │ │
│  │ └─────────────────────────────────────────────────┘   │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                          Workspace                           │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ packagesFile: .kloudlite/packages.yaml                 │ │
│  └────────────────────────────────────────────────────────┘ │
│                             │                                │
│                             ▼                                │
│  ┌────────────────────────────────────────────────────────┐ │
│  │            PackageRequest CR                           │ │
│  │  Spec:                                                 │ │
│  │    - nodeName: <workmachine-node>                      │ │
│  │    - packages: [jq, curl, git, nodejs]                 │ │
│  │  Status:                                               │ │
│  │    - phase: Ready                                      │ │
│  │    - installedPackages: [...]                          │ │
│  └────────────────────────────────────────────────────────┘ │
│                             │                                │
│                             ▼                                │
│  ┌────────────────────────────────────────────────────────┐ │
│  │            Workspace Pod                               │ │
│  │  - Mounts nix-store from host (read-only)             │ │
│  │  - PATH includes package binaries                      │ │
│  │  - Only starts after PackageRequest is Ready           │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Key Files

### Implementation
- `api/pkg/apis/packages/v1/packagerequest_types.go` - CRD type definitions
- `api/cmd/workmachine-node-manager/main.go` - DaemonSet controller application
- `api/Dockerfile.workmachine-node-manager` - Container image definition
- `api/internal/controllers/workmachine_controller.go:103-109` - DaemonSet deployment
- `api/internal/controllers/workspace_controller.go` - PackageRequest creation

### Manifests
- `devenv/manifests/packages.kloudlite.io_packagerequests.yaml` - CRD manifest
- `devenv/manifests/workmachine-node-manager-rbac.yaml` - RBAC resources

## 🎯 How It Works

1. **WorkMachine starts**: Creates workmachine-node-manager Deployment in WorkMachine's targetNamespace
2. **Deployment pod starts**: Watches PackageRequest CRs in its namespace
3. **Workspace created**: Controller reads `.kloudlite/packages.yaml` from workspace repo
4. **PackageRequest created**: Workspace controller creates CR with package list in the namespace
5. **Packages installed**: WorkMachine node manager pod sees the request, installs packages to `/var/lib/kloudlite/nix-store`
6. **Status updated**: PackageRequest status updated to "Ready" with installed package info
7. **Workspace pod starts**: Only after PackageRequest is Ready, mounts nix-store and uses packages

## 📝 Example packages.yaml

```yaml
packages:
  # CLI tools
  - name: jq
  - name: yq
  - name: curl
  - name: wget

  # Development tools
  - name: git
  - name: vim
  - name: tmux

  # Programming languages
  - name: nodejs
  - name: python3
  - name: go

  # Build tools
  - name: make
  - name: cmake
```

## 🚀 Quick Start Alias

Add this to your shell config for easier kubectl access:

```bash
alias k='kubectl --kubeconfig=/Users/karthik/dev/kl-workspace/kloudlite-v2/devenv/k3s-config/k3s.yaml'
```

Then use:
```bash
k get packagerequest -A
k describe packagerequest <name> -n <namespace>
k logs -f daemonset/package-cache -n default
```
