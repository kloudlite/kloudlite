# kl - Kloudlite Workspace Manager

A powerful command-line tool for managing Kloudlite workspaces from inside the workspace container.

## Overview

`kl` is a CLI binary that runs inside workspace pods and provides an intuitive interface to manage workspace resources via the Kubernetes API. It allows users to view workspace status, install/uninstall Nix packages, and configure workspace settings.

Built with [Cobra](https://github.com/spf13/cobra), it features command aliases, rich help text, and an intuitive command structure for improved developer experience.

## Requirements

- Runs inside a Kubernetes pod with access to the workspace Custom Resource
- Requires the following environment variables:
  - `WORKSPACE_NAME`: Name of the workspace resource (defaults to `HOSTNAME`)
  - `WORKSPACE_NAMESPACE`: Namespace of the workspace resource (defaults to `default`)
- Requires appropriate RBAC permissions to read and update Workspace CRDs

## Commands

All commands support short aliases for faster typing. Use `kl <command> --help` for detailed information about any command.

### Status

Display comprehensive workspace information:

```bash
kl status    # Full command
kl st        # Short alias
kl s         # Shortest alias
```

Shows:
- Workspace metadata (name, owner, display name)
- Current phase and status
- Resource usage (CPU, memory, storage)
- Resource quotas
- Access URLs
- Timing information (start time, last activity, total runtime)
- Active connections

### Package Management

All package commands support aliases: `pkg`, `package`, or `p`.

#### Search for packages

Search the Devbox package registry:

```bash
kl pkg search nodejs    # Full command
kl p s python          # With aliases
kl p find vim          # Alternative alias
```

#### Add packages

Add packages interactively with fuzzy search and version selection:

```bash
kl pkg add    # Full command
kl p a        # With aliases (interactive mode)
```

Add one or more packages directly by name (uses latest version):

```bash
kl pkg add git vim curl    # Full command
kl p a nodejs python       # With aliases
```

#### Install a Nix package

Install with specific version, channel, or commit:

```bash
# Latest version
kl pkg install nodejs
kl p i python             # With aliases

# Specific version
kl pkg install nodejs --version 20.0.0
kl p i python --version 3.11.0

# From channel
kl pkg install vim --channel nixos-24.05
kl p i git --channel unstable

# From specific commit
kl pkg install curl --commit abc123def456
```

Options:
- `--version`: Semantic version (e.g., `24.0.0`)
- `--channel`: Nixpkgs channel (e.g., `nixos-24.05`, `nixos-23.11`, `unstable`)
- `--commit`: Exact nixpkgs commit hash

#### Uninstall a package

Remove packages from workspace:

```bash
kl pkg uninstall git    # Full command
kl p rm vim            # With aliases
kl p un nodejs         # Alternative alias
```

#### List packages

Display all packages and their status:

```bash
kl pkg list    # Full command
kl p ls        # Short alias
kl p l         # Shortest alias
```

Shows:
- Packages in workspace spec
- Installed packages with version, binary path, and installation time
- Failed packages with error messages

### Configuration Management

All config commands support aliases: `config`, `cfg`, or `c`.

#### View configuration

View all workspace configuration:

```bash
kl config get    # Full command
kl c get         # With alias
```

View specific configuration value:

```bash
kl config get display-name
kl c get git.user-email    # With alias
kl c get env.NODE_ENV
```

Supported keys:
- `display-name`
- `description`
- `owner`
- `storage-size`
- `workspace-path`
- `vscode-version`
- `auto-stop`
- `idle-timeout`
- `max-runtime`
- `git.user-name`
- `git.user-email`
- `git.default-branch`
- `env.<VAR_NAME>` - for environment variables

#### Update configuration

Update any configuration value:

```bash
kl config set <key> <value>    # Full command
kl c set <key> <value>          # With alias
```

Examples:

```bash
kl config set display-name "My Dev Workspace"
kl c set description "Development environment for project X"
kl c set auto-stop true
kl c set git.user-name "John Doe"
kl c set git.user-email "john@example.com"
kl c set env.NODE_ENV production
```

### Version

Display CLI version:

```bash
kl version    # Full command
kl v          # With alias
```

### Help

Display usage information for any command:

```bash
kl help
kl --help
kl -h

# Get help for specific commands
kl pkg --help
kl p add --help
kl config set --help
```

## Building

The binary is intended to run on Linux (inside containers).

### Using Taskfile (Recommended)

From the `devenv/` directory:

```bash
# Build and install to k3s
task kl:build-install
```

### Manual Build

```bash
GOOS=linux GOARCH=amd64 go build -o kl
```

## Integration

The `kl` binary is automatically made available in workspace pods:

1. **Build & Install**: Use `task kl:build-install` to build and copy the binary to k3s
2. **Auto-mount**: The workspace controller automatically mounts `/kloudlite/bin/kl` to `/usr/local/bin/kl` in each workspace pod
3. **Environment Variables**: The controller injects required environment variables (`WORKSPACE_NAME`, `WORKSPACE_NAMESPACE`)
4. **RBAC**: The workspace service account needs appropriate permissions (see below)

### Required RBAC

The service account used by the workspace pod needs:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: workspace-user
rules:
- apiGroups: ["workspaces.kloudlite.io"]
  resources: ["workspaces"]
  verbs: ["get", "update", "patch"]
```

## Architecture

The CLI uses:
- `controller-runtime/pkg/client` for Kubernetes API interactions
- In-cluster config when running in a pod
- Falls back to local kubeconfig for development/testing
- Direct updates to Workspace CRD specs via Kubernetes API
- JSON merge patches to avoid update conflicts
