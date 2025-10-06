# Kloudlite Workspace Images

Declarative development environments with Nix package manager, supporting multiple editors and servers.

## Overview

This directory contains base images and server variants for Kloudlite workspaces with declarative package management using Nix.

## Architecture

```
workspace-images/
├── base/                  # Base image with Nix
├── servers/               # Server variants
│   ├── code-server/       # VS Code in browser
│   ├── jupyter/           # JupyterLab
│   ├── ttyd/              # Web terminal
│   └── code-web/          # Monaco editor
└── examples/              # Example Nix configurations
```

## Server Variants

### code-server
VS Code running in the browser with full IDE capabilities.
- **Port**: 8080
- **Image**: `kloudlite/workspace-code-server:latest`

### jupyter
JupyterLab for interactive notebooks and data science.
- **Port**: 8888
- **Image**: `kloudlite/workspace-jupyter:latest`

### ttyd
Web-based terminal for command-line access.
- **Port**: 7681
- **Image**: `kloudlite/workspace-ttyd:latest`

### code-web
Lightweight Monaco editor for quick edits.
- **Port**: 3000
- **Image**: `kloudlite/workspace-code-web:latest`

## Declarative Package Management

### Using YAML Configuration

Create a `.kloudlite/packages.yaml` file in your project:

```yaml
packages:
  - nodejs-18_x
  - python311
  - go
  # Add any packages you need
```

The workspace will automatically install these packages on startup using Nix package manager.

### Example Configurations

See `examples/` directory for pre-made configurations:
- `python-ml.yaml` - Python + ML/DS libraries
- `nodejs-web.yaml` - Node.js + web development tools
- `rust-dev.yaml` - Rust toolchain + build tools

## Building Images

```bash
# Build base image
cd base
docker build -t kloudlite/workspace-base:latest .

# Build server variant
cd servers/code-server
docker build -t kloudlite/workspace-code-server:latest .
```

## Usage in Workspace CRD

```yaml
apiVersion: workspaces.kloudlite.io/v1
kind: Workspace
metadata:
  name: my-workspace
spec:
  displayName: "My Development Workspace"
  owner: "user@example.com"
  hostPath: "/path/to/project"
  serverType: "code-server"  # or jupyter, ttyd, code-web
  packagesFile: ".kloudlite/packages.yaml"  # optional
```

## Supported Package Managers (via Nix)

Nix provides access to virtually all package managers and tools:
- **Languages**: Python (pip), Node.js (npm/yarn/pnpm), Rust (cargo), Go, Ruby (gem), PHP (composer)
- **System**: apt packages, build tools (gcc, make, cmake)
- **Databases**: PostgreSQL, MySQL, MongoDB, Redis
- **Tools**: Docker, kubectl, terraform, and 80,000+ packages

## Environment Variables

- `WORKSPACE_PACKAGES_FILE`: Path to packages YAML file (default: `/workspace/.kloudlite/packages.yaml`)
- `STARTUP_SCRIPT`: Path to startup script to run after package installation

## Benefits

- **Declarative**: Define your entire environment in code
- **Reproducible**: Same config = same environment everywhere
- **Version-controlled**: Commit your environment configuration
- **Multi-language**: Support any language or tool via Nix
- **Fast**: Binary caches make installation quick
- **Isolated**: Each workspace gets its own package set

## Troubleshooting

### Packages not installing
- Check YAML syntax in your `packages.yaml` file
- View logs: `kubectl logs <workspace-pod>`
- Verify package names at https://search.nixos.org

### Slow first startup
- First build downloads packages; subsequent starts use cache
- Enable binary caches (already configured in base image)

## Documentation

- Nix packages: https://search.nixos.org/packages
- Nix language: https://nixos.org/manual/nix/stable/language/
- Examples: See `examples/` directory
