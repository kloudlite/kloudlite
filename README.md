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

**Kloudlite** gives you instant cloud development environments connected to your production-like infrastructure. Write code, intercept live services, and test changes in real time — no build or deploy steps required.

## Why Kloudlite?

**The Problem**: Traditional development loops are slow. You write code, build, deploy to staging, wait, test, find bugs, and repeat. Each cycle takes minutes to hours.

**The Solution**: Kloudlite eliminates the loop. Your workspace is already connected to your environment. Intercept any service, and traffic flows directly to your code. Changes are instant.

## Features

### Cloud Workspaces
Get a full development environment in seconds. Access via:
- **SSH** — Use your favorite local IDE with remote SSH
- **VS Code** — Browser-based VS Code with full extension support
- **Web Terminal** — Quick access from any browser

### Service Interception
Intercept any service running in your environment. Traffic that would go to that service comes to your workspace instead. Debug with real requests, test your changes instantly.

```
Production traffic → Your intercepted service → Your workspace code
```

### Environment Connections
Connect your workspace to any environment. All services become accessible — databases, APIs, queues. DNS just works.

### Package Management
Install any package instantly with Nix:
```bash
kl pkg add nodejs python3 go
```

### Port Exposure
Expose any port from your workspace with a public URL. Perfect for:
- Testing webhooks
- Sharing work with teammates
- Mobile app development

### AI-Powered Development
Built-in MCP server for AI coding assistants. Works with Claude, Codex, and OpenCode out of the box.

### Local Access
Connect your local machine via `kltun` to access all workspace and environment services directly from your local IDE. Your local tools, your cloud environment.

## Getting Started

1. **Create a workspace** from the [Kloudlite Dashboard](https://kloudlite.io)
2. **Connect** via SSH, VS Code, or web terminal
3. **Link to an environment** to access services
4. **Start coding** — intercept services, expose ports, install packages

## Resources

- [Website](https://kloudlite.io)
- [Documentation](https://kloudlite.io/docs)
- [Discord Community](https://discord.com/invite/m5tYzQfcG8)
- [Twitter](https://x.com/kloudlite)

## Security

Found a vulnerability? Report it to support@kloudlite.io.

## License

AGPL Version 3.0 — See [LICENSE](LICENSE) for details.
