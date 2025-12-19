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

Kloudlite provides secure, instant-access cloud development environments that connect directly to your production-like infrastructure. Eliminate the traditional build-deploy-test cycle by working in environments that mirror production, with the ability to intercept live services and test changes in real time.

## The Problem

Traditional development workflows are slow:

1. Write code locally
2. Build and deploy to staging
3. Wait for deployment
4. Test and discover issues
5. Repeat

Each iteration takes minutes to hours. Developers spend more time waiting than coding.

## The Solution

With Kloudlite, your development environment is already connected to your infrastructure:

- **No deployment required** — Your code runs where your services run
- **Real traffic testing** — Intercept production services and debug with actual requests
- **Instant feedback** — See changes immediately without build or deploy steps

## Core Capabilities

### Secure Access

Connect to your workspace and environment services through an encrypted tunnel using `kltun`. All traffic between your local machine and the cloud environment is secured, giving you direct access to every service as if you were on the same network.

### Cloud Workspaces

Full-featured development environments accessible via:

- **SSH** — Connect with your preferred local IDE
- **VS Code** — Browser-based IDE with complete extension support
- **Web Terminal** — Instant access from any browser

### Service Interception

Redirect traffic from any running service to your workspace. Incoming requests flow directly to your local code, enabling real-time debugging with production traffic patterns.

### Environment Connections

Connect your workspace to any environment and gain immediate access to all services — databases, APIs, message queues, and internal tools. DNS resolution is automatic.

### Package Management

Install development tools and dependencies instantly:

```bash
kl pkg add nodejs python3 postgresql redis
```

Powered by Nix for reproducible, conflict-free package management.

### Port Exposure

Expose workspace ports with secure public URLs for webhook testing, mobile development, or sharing work with teammates.

### AI Assistant Integration

Built-in MCP (Model Context Protocol) server enables AI coding assistants to interact directly with your workspace. Pre-configured for Claude, Codex, and OpenCode.

## Getting Started

### For Platform Administrators

1. **Deploy Kloudlite** on your cloud provider (AWS, GCP, Azure, or self-hosted)
2. **Configure** work machines and resource quotas
3. **Invite** team members and set up access controls

### For Developers

1. **Connect** to your work machine using the provided `kltun` command
2. **Create** your workspace and environments
3. **Start building** — access services, intercept traffic, collaborate with your team

## Documentation

- [Getting Started Guide](https://kloudlite.io/docs)
- [CLI Reference](https://kloudlite.io/docs/cli)
- [API Documentation](https://kloudlite.io/docs/api)

## Community

- [Discord](https://discord.com/invite/m5tYzQfcG8) — Get help and connect with other developers
- [Twitter](https://x.com/kloudlite) — Latest updates and announcements
- [GitHub Issues](https://github.com/kloudlite/kloudlite/issues) — Report bugs and request features

## Security

We take security seriously. If you discover a vulnerability, please report it to **security@kloudlite.io**.

## License

Kloudlite is open source under the [AGPL-3.0 License](LICENSE).

---

<div align="center">
  <strong>Stop waiting. Start building.</strong>
  <br/><br/>
  <a href="https://kloudlite.io">Get Started →</a>
</div>
