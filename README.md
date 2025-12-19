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

Kloudlite is an open-source platform delivering secure, production-parity development environments for engineering and QA teams. Intercept live services, switch environments on the fly, and validate changes against real infrastructure — without waiting for builds or deployments.

## Why Kloudlite?

### The Problem

Engineering teams burn hours on feedback loops:

- Code → Build → Deploy → Wait → Test → Debug → Repeat
- QA can't reproduce issues without matching environment configurations
- Environment sharing requires manual setup and constant synchronization
- Context switching between environments kills productivity

### The Solution

Kloudlite collapses the inner development loop:

- **Zero deployment testing** — Run your code against live services instantly
- **Environment agility** — Hot-swap between dev, staging, and prod-mirror environments
- **Parallel collaboration** — Engineers and QA operate in shared environments simultaneously
- **Parity guarantee** — Identical environments across development and testing workflows

## Core Capabilities

### Secure Tunnel

Encrypted tunnel via `kltun` establishes direct connectivity to your work machine. One command gives you network-level access to all services in your connected environments.

### Cloud Workspaces

Production-grade development environments:

- **SSH** — Remote development with your local IDE
- **VS Code Server** — Full IDE in browser with extension support
- **Web Terminal** — Direct shell access via ttyd

### Multi-Environment Switching

Bind your workspace to any environment — development, staging, feature-branch, or prod-mirror. Switch contexts without rebuilding. Each environment maintains its own service topology, databases, and configurations.

### Team Collaboration

Multiple engineers and QA connect to the same environment concurrently. Debug together, run parallel test suites, or validate the same fix simultaneously — isolated workspaces, shared infrastructure.

### Service Interception

Route traffic from any service directly to your workspace. Intercept requests at the service mesh level, debug with production traffic patterns, and validate fixes before they hit CI/CD.

### Environment Connectivity

Full service discovery within connected environments. Access databases, APIs, queues, caches — all resolvable by service name. No port-forwarding, no proxy configuration.

### Package Management

Nix-powered package installation:

```bash
kl pkg add go@1.21 nodejs python3 postgresql-client
```

Reproducible, isolated, no dependency conflicts.

### Port Exposure

Expose local ports with public endpoints:

- Webhook receivers
- OAuth callbacks
- Mobile backend testing
- Stakeholder demos

### AI Tooling

Native MCP server integration. Claude, Codex, and OpenCode can manage packages, switch environments, and control intercepts directly from your conversation.

## Use Cases

### Development

Test against real services without deployment. Switch between feature environments instantly. Pair program in shared workspaces.

### QA & Testing

Access identical environments as engineering. Reproduce issues reliably. Validate fixes in real time without waiting for release cycles.

### Platform Engineering

Single deployment across cloud providers. Centralized work machine and environment management. Fine-grained access controls per team.

## Getting Started

### Infrastructure Setup

1. Deploy Kloudlite on your cloud (AWS, GCP, Azure) or self-host
2. Provision work machines and set resource quotas
3. Define environments mirroring your infrastructure
4. Configure team access and permissions

### Developer Workflow

1. Run `kltun` to establish secure tunnel
2. Create workspace and bind to target environment
3. Intercept services as needed
4. Switch environments without workspace recreation

## Resources

- [Documentation](https://kloudlite.io/docs)
- [CLI Reference](https://kloudlite.io/docs/cli)
- [API Docs](https://kloudlite.io/docs/api)
- [Helm Charts](https://github.com/kloudlite/helm-charts)

## Community

- [Discord](https://discord.com/invite/m5tYzQfcG8)
- [Twitter](https://x.com/kloudlite)
- [GitHub Issues](https://github.com/kloudlite/kloudlite/issues)

## Security

All traffic encrypted. Environments isolated. Access controlled. Report vulnerabilities to **security@kloudlite.io**.

## License

[AGPL-3.0](LICENSE)

---

<div align="center">
  <strong>Faster feedback. Seamless collaboration. Ship with confidence.</strong>
  <br/><br/>
  <a href="https://kloudlite.io">Get Started →</a>
</div>
