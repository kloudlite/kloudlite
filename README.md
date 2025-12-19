<div align="center">
  <a href="https://kloudlite.io">
    <img src="https://github.com/kloudlite/kloudlite/assets/1580519/a31a5f78-2bde-45f1-8141-d23ee8231eb1" style="height:38px" />
  </a>
  <h1>Kloudlite</h1>
  <p><strong>Cloud Development Environments</strong></p>
  <p>Reduce your development loop from minutes to seconds</p>

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

## Overview

Kloudlite is an open-source platform that provides secure, production-mirror cloud development environments for developers and QA teams. Connect to any environment, intercept live services, and test changes instantly — eliminating the traditional build-deploy-wait cycle that slows down software delivery.

## Why Kloudlite?

### The Problem

Software teams lose countless hours to slow feedback loops:

- Developers wait for builds and deployments to test changes
- QA teams struggle to reproduce issues without access to the right environment
- Collaboration requires complex environment sharing and configuration
- Switching between environments means rebuilding context every time

### The Solution

Kloudlite eliminates these barriers:

- **Instant feedback** — Test changes in real time without deploying
- **Environment switching** — Move between dev, staging, and production-like environments seamlessly
- **Team collaboration** — Multiple team members connect to the same environment simultaneously
- **Unified platform** — Developers and QA work in identical environments, eliminating "works on my machine"

## Core Capabilities

### Secure Access

All connections are secured through encrypted tunnels. Run a single command to establish a secure connection to your work machine and gain direct access to every service in your connected environments.

### Cloud Workspaces

Fully configured development environments ready in seconds:

- **SSH** — Use your preferred local IDE with remote development
- **VS Code** — Browser-based IDE with full extension support
- **Web Terminal** — Instant access from any browser, anywhere

### Multi-Environment Support

Switch between environments instantly. Connect your workspace to development, staging, or any custom environment — each with its own services, databases, and configurations. No rebuild, no redeploy, just switch and continue working.

### Real-Time Collaboration

Multiple developers and QA engineers can connect to the same environment simultaneously. Share your workspace, debug together, or run parallel tests — all without stepping on each other's work.

### Service Interception

Intercept any service and redirect its traffic to your workspace. Debug with real production traffic patterns, test edge cases with actual requests, and validate fixes before deployment.

### Environment Connections

Connect to any environment and instantly access all its services — databases, APIs, message queues, caches, and internal tools. DNS resolution is automatic; services are accessible by name.

### Package Management

Install any development tool instantly:

```bash
kl pkg add nodejs python3 postgresql redis
```

Powered by Nix for reproducible, isolated, and conflict-free package management.

### Port Exposure

Expose workspace ports with secure public URLs. Essential for:

- Webhook development and testing
- Mobile app backend development
- Sharing work-in-progress with stakeholders
- QA testing of unreleased features

### AI Assistant Integration

Built-in MCP server enables AI coding assistants to manage packages, control environment connections, and handle service interception directly. Pre-configured for Claude, Codex, and OpenCode.

## Who Is It For?

### Developers

Write code and test it instantly against real services. Switch between feature branches and environments without losing context. Collaborate with teammates in shared environments.

### QA Engineers

Access the exact same environments as developers. Reproduce bugs reliably, test against real service configurations, and validate fixes in real time — no more waiting for deployments.

### Platform Teams

Deploy once on your cloud provider. Manage work machines, environments, and access controls centrally. Give your teams the infrastructure they need without the overhead.

## Getting Started

### Platform Setup

1. **Deploy** Kloudlite on your cloud provider (AWS, GCP, Azure, or self-hosted)
2. **Configure** work machines and define resource quotas
3. **Create** environments that mirror your infrastructure
4. **Invite** team members and configure access controls

### Developer & QA Workflow

1. **Connect** to your work machine using the provided connection command
2. **Create** your workspace
3. **Switch** between environments as needed
4. **Collaborate** with your team in shared environments
5. **Intercept** services to test and debug with real traffic

## Documentation

- [Getting Started Guide](https://kloudlite.io/docs)
- [CLI Reference](https://kloudlite.io/docs/cli)
- [API Documentation](https://kloudlite.io/docs/api)
- [Helm Charts](https://github.com/kloudlite/helm-charts)

## Community

- [Discord](https://discord.com/invite/m5tYzQfcG8) — Get help and connect with other users
- [Twitter](https://x.com/kloudlite) — Latest updates and announcements
- [GitHub Issues](https://github.com/kloudlite/kloudlite/issues) — Report bugs and request features

## Security

Security is foundational to Kloudlite. All connections are encrypted, environments are isolated, and access is controlled. If you discover a vulnerability, please report it to **security@kloudlite.io**.

## License

Kloudlite is open source under the [AGPL-3.0 License](LICENSE).

---

<div align="center">
  <strong>Faster feedback. Better collaboration. Ship with confidence.</strong>
  <br/><br/>
  <a href="https://kloudlite.io">Get Started →</a>
</div>
