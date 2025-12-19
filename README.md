<div align="center">
  <a href="https://kloudlite.io">
    <img src="https://github.com/kloudlite/kloudlite/assets/1580519/a31a5f78-2bde-45f1-8141-d23ee8231eb1" style="height:38px" />
  </a>
  <h1>Kloudlite</h1>
  <p><strong>Cloud Development Environments</strong></p>
  <p>Test against live services without deploying</p>

  <a href="https://discord.gg/m5tYzQfcG8">
    <img src="https://img.shields.io/discord/934762910717194260?label=discord" alt="Discord">
  </a>
  <a href="#license">
    <img src="https://img.shields.io/github/license/kloudlite/kloudlite" alt="License">
  </a>
  <a href="https://goreportcard.com/report/github.com/kloudlite/api">
    <img src="https://goreportcard.com/badge/github.com/kloudlite/api" alt="Go Report Card">
  </a>
  <a href="https://github.com/kloudlite/kloudlite/issues">
    <img src="https://img.shields.io/github/issues/kloudlite/kloudlite" alt="Issues">
  </a>
  <a href="https://github.com/kloudlite/kloudlite/releases">
    <img src="https://img.shields.io/github/v/release/kloudlite/kloudlite" alt="GitHub Release">
  </a>
</div>

<br/>

## What is Kloudlite?

Kloudlite provides cloud-based development workspaces with live service connectivity. Think Telepresence meets cloud IDEs — but with per-developer environment ownership, instant environment switching, and cross-team collaboration built in.

Your code runs against real services. No container builds. No deployments. No waiting.

## The Inner Loop Problem

The traditional development cycle — code, build, deploy, test — takes minutes per iteration. Most of that time is spent waiting for builds and deployments, not actually validating changes.

Kloudlite eliminates build and deploy from your inner loop. You write code in a cloud workspace that's already connected to your services. Changes are testable immediately. The feedback loop drops from minutes to seconds.

## Core Concepts

**Workspace** — A container running on your work machine with your dev tools installed. Accessible via SSH, VS Code Server, or web terminal. Mounts code from the host filesystem. Multiple workspaces share system volumes for efficiency.

**Environment** — An isolated namespace containing your services, databases, and configurations. You own your environments — create as many as you need for different features or experiments. Switch your workspace between them instantly.

**Intercept** — Routes traffic from any service in your environment to your workspace. Debug with production-like traffic patterns. Validate fixes before pushing.

**Tunnel** — WireGuard-based connection from your local machine to your work machine via `kltun`. All services become DNS-resolvable locally. Your IDE connects over this tunnel.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                         Kubernetes Cluster (Team)                            │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │                          Control Plane                                 │  │
│  │     API Server  ◄───►  Dashboard  ◄───►  Kubernetes Controllers        │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌──────────────────────────┐      ┌──────────────────────────┐             │
│  │   Work Machine (Dev A)   │      │   Work Machine (Dev B)   │    ...      │
│  │                          │      │                          │             │
│  │  ┌────────────────────┐  │      │  ┌────────────────────┐  │             │
│  │  │  Namespace: dev-a  │  │      │  │  Namespace: dev-b  │  │             │
│  │  │   ┌─────────────┐  │  │      │  │   ┌─────────────┐  │  │             │
│  │  │   │ Workspace 1 │  │  │      │  │   │ Workspace 1 │  │  │             │
│  │  │   │ Workspace 2 │  │  │      │  │   │ Workspace 2 │  │  │             │
│  │  │   └─────────────┘  │  │      │  │   └─────────────┘  │  │             │
│  │  └────────────────────┘  │      │  └────────────────────┘  │             │
│  │                          │      │                          │             │
│  │  ┌────────────────────┐  │      │  ┌────────────────────┐  │             │
│  │  │ Env: dev-a-feature │  │      │  │ Env: dev-b-feature │  │             │
│  │  │  (services, DBs)   │  │      │  │  (services, DBs)   │  │             │
│  │  └────────────────────┘  │      │  └────────────────────┘  │             │
│  └──────────────────────────┘      └──────────────────────────┘             │
│                                                                              │
│        All nodes can reach any environment — collaborate across teams        │
└──────────────────────────────────────────────────────────────────────────────┘
           ▲                                        ▲
           │ WireGuard (kltun)                      │ WireGuard (kltun)
           ▼                                        ▼
   ┌───────────────┐                        ┌───────────────┐
   │  Dev A Local  │                        │  Dev B Local  │
   │   (IDE, CLI)  │                        │   (IDE, CLI)  │
   └───────────────┘                        └───────────────┘
```

**Work Machine** — Dedicated node per developer in the cluster. Runs your workspaces and environments.

**Workspaces** — Containers in your namespace. Share host volumes for code and system packages. Include:
- Nix package management (`kl pkg add go@1.21 nodejs python3`)
- IDE integration (SSH, VS Code Server, Web Terminal)
- Mounted workspace folders from host

**Environments** — Isolated namespaces with your services. Each developer owns their environments. Connect to any team member's environment for debugging or collaboration.

## Capabilities

| Feature | Description |
|---------|-------------|
| **Service Interception** | Route traffic from any service to your workspace for live debugging |
| **Multi-Environment** | Own multiple environments, switch between them without rebuilds |
| **Cross-Team Access** | Connect to any environment in the cluster for collaboration |
| **Nix Packages** | `kl pkg add go@1.21 nodejs` — reproducible, conflict-free |
| **Port Exposure** | Public URLs for webhooks, OAuth callbacks, external testing |
| **MCP Server** | AI tools (Claude, Codex, OpenCode) control workspace via `kl mcp` |

## Project Structure

```
api/
├── cmd/server/                    # Control plane API server
├── cmd/kl/                        # CLI (runs inside workspace)
├── cmd/workmachine-node-manager/  # Host-level Nix package management
├── internal/controllers/          # K8s controllers
│   ├── workspace/                 # Workspace lifecycle
│   ├── environment/               # Environment management
│   └── serviceintercept/          # Traffic interception
└── manifests/                     # CRDs and RBAC

web/                               # Next.js dashboard
devenv/                            # Local K3s development setup
```

**Stack:** Go 1.24, controller-runtime, Kubernetes CRDs, WireGuard, Nix, Next.js 15, React 19

## Getting Started

- **Self-host:** Deploy on AWS, GCP, or Azure — [Installation Guide](https://kloudlite.io/docs/install)
- **Documentation:** [kloudlite.io/docs](https://kloudlite.io/docs)
- **CLI Reference:** [kloudlite.io/docs/cli](https://kloudlite.io/docs/cli)
- **Helm Charts:** [github.com/kloudlite/helm-charts](https://github.com/kloudlite/helm-charts)

## Community

- [Discord](https://discord.com/invite/m5tYzQfcG8)
- [Twitter](https://x.com/kloudlite)
- [GitHub Issues](https://github.com/kloudlite/kloudlite/issues)

Security issues: **security@kloudlite.io**

## License

[AGPL-3.0](LICENSE)
