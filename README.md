<div align="center">
  <a href="https://kloudlite.io">
    <img src="https://github.com/kloudlite/kloudlite/assets/1580519/a31a5f78-2bde-45f1-8141-d23ee8231eb1" style="height:38px" />
  </a>
  <h1>
    Cloud Development Environments
  </h1>
  <p><strong>Designed to reduce the development loop</strong></p>

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

**Kloudlite** is an open-source platform that provides cloud development environments with instant access to production-like infrastructure. Skip the build-deploy cycle — connect your workspace to environments, intercept services, and see your changes in real time.

## Key Features

- **Cloud Workspaces** — Full development environments accessible via SSH, VS Code, or web terminal (ttyd)
- **Service Interception** — Redirect traffic from any environment service to your workspace for real-time debugging
- **Environment Connections** — Connect workspaces to environments with automatic DNS resolution for all services
- **Nix Package Management** — Install any package instantly with `kl pkg add`
- **Port Exposure** — Expose workspace ports with public URLs for webhooks and sharing
- **AI Tool Integration** — Built-in MCP server for Claude, Codex, and OpenCode
- **Local VPN Access** — Connect your local machine to the cluster via `kltun` for direct service access

## How It Works

Kloudlite runs workspaces and environments in the same Kubernetes cluster:

```
┌─────────────────────────────────────────────────────────────┐
│                    Kloudlite Cluster                        │
│                                                             │
│  ┌─────────────────┐       ┌─────────────────────────────┐  │
│  │    Workspace    │       │        Environment          │  │
│  │                 │       │                             │  │
│  │  • VS Code      │◄─────►│  • Services                 │  │
│  │  • SSH Access   │ intercept │  • Deployments          │  │
│  │  • Web Terminal │       │  • Databases                │  │
│  │  • kl CLI       │       │                             │  │
│  └─────────────────┘       └─────────────────────────────┘  │
│           ▲                                                 │
└───────────│─────────────────────────────────────────────────┘
            │ WireGuard VPN (kltun)
            │
    ┌───────┴───────┐
    │ Local Machine │
    │   (optional)  │
    └───────────────┘
```

**Service Intercepts**: When you intercept a service, traffic destined for that service is redirected to your workspace using SOCAT port forwarding. This lets you debug with real production traffic without deploying.

**Local VPN (`kltun`)**: Optionally connect your local machine directly to the cluster via WireGuard VPN for accessing services and workspaces from your local IDE.

## Components

| Component | Description |
|-----------|-------------|
| **Dashboard** | Web UI for managing workspaces and environments |
| **kl** | CLI tool inside workspaces for package management and configuration |
| **kltun** | VPN client for connecting local machines to the cluster |
| **Controllers** | Kubernetes operators managing workspace and environment lifecycle |

## Resources

- [Website](https://kloudlite.io)
- [Documentation](https://kloudlite.io/docs)
- [Roadmap](https://github.com/orgs/kloudlite/projects/22/views/5)
- [Helm Charts](https://github.com/kloudlite/helm-charts)

## Contact

- [Discord](https://discord.com/invite/m5tYzQfcG8) — Join our community
- [Twitter](https://x.com/kloudlite) — Follow us
- [Contact Us](https://kloudlite.io/contact-us)

## Security

If you've found a security vulnerability, please report it to support@kloudlite.io.

## License

Kloudlite is distributed under the AGPL Version 3.0 license. See [LICENSE](LICENSE) for details.

## Built With

- [Kubernetes](https://github.com/kubernetes/kubernetes) & [K3s](https://github.com/k3s-io/k3s) — Container orchestration
- [WireGuard](https://github.com/WireGuard) — VPN for local machine connectivity
- [Nix](https://github.com/NixOS/nix) — Reproducible package management
- [Next.js](https://github.com/vercel/next.js) — Web dashboard
- [Go](https://github.com/golang/go) — Backend services and CLI
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) — Kubernetes controllers

