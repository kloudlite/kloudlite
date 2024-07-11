<div align="center">
  <a href="https://kloudlite.io">
    <img src="https://github.com/kloudlite/kloudlite/assets/1580519/a31a5f78-2bde-45f1-8141-d23ee8231eb1" style="height:38px" />
  </a>
  <h1>
    Development Environment as a Service
  </h1>
  <div>No more waiting for Build & Deploy. Just code & run!</div>
  <br />
  
  <a href="https://kloudlite.io">Quickstart</a> |
  <a href="https://kloudlite.io/docs">Docs</a> |
  <a href="https://kloudlite.io/why">Why Kloudlite?</a> |
  <a href="https://kloudlite.io/install">Install</a>
  <br />

  <a href="https://discord.gg/m5tYzQfcG8">
    <img src="https://img.shields.io/discord/934762910717194260?label=discord" alt="Discord">
  </a>
  <a href="#license">
    <img src="https://img.shields.io/badge/License-Apache--2.0-blue" alt="License">
  </a>
  <a href="https://goreportcard.com/report/github.com/kloudlite/api">
    <img src="https://goreportcard.com/badge/github.com/kloudlite/api" alt="Go Report Card">
  </a>
  <a href="https://github.com/kloudlite/kloudlite/issues">
    <img src="https://img.shields.io/github/issues/kloudlite/kloudlite" alt="Issues">
  </a>
  <img src="https://img.shields.io/github/v/release/kloudlite/kloudlite" alt="GitHub Release">
  <a href="https://console.algora.io/org/kloudlite/bounties?status=open">
    <img src="https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fkloudlite%2Fbounties%3Fstatus%3Dopen" alt="Open Bounties">
  </a>
  <a href="https://console.algora.io/org/kloudlite/bounties?status=completed">
    <img src="https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fkloudlite%2Fbounties%3Fstatus%3Dcompleted" alt="Rewarded Bounties">
  </a>
</div>

Kloudlite is a platform designed to enhance developers' productivity by providing seamless, secure, **production-parity development environments**. It connects local systems and remote environments via a WireGuard network, allowing developers to build, test, and deploy distributed applications efficiently. Kloudlite eliminates the need for separate configurations by syncing configurations and secrets across environments. It supports collaborative coding, real-time testing, and debugging.

## Quickstart
The quickest way to start using Kloudlite is to use our hosted solution. [Login](https://auth.kloudlite.io) and set up your Kloudlite account.

### Attach Cluster
Attach your cluster in the infrastructure section of the Kloudlite dashboard.

![Frame 2](https://github.com/kloudlite/kloudlite/assets/1580519/e9629f43-0d44-4311-b6f4-e02265ec7d3b)

### Set Up Environment
Create your environment. Add apps, configs, and secrets to your environment.

![Frame 4](https://github.com/kloudlite/kloudlite/assets/1580519/9bb3c9ec-5c25-4a99-b038-9f17b0c40710)

### Install Kloudlite CLI
```bash
# Setup and run docker on your machine

# Install kl
curl 'https://kl.kloudlite.io/kloudlite/kl!?select=kl' | bash

# Login with kloudlite
kl auth login

```

### Access Environment
```bash
cd workspace

# Setup workspace. You will be asked to choose the team and default environment of the current workspace.
kl init

# SSH into the development container and access
kl box ssh
```

## Install on your own cloud
There are three components of Kloudlite that need to be installed to run:
- Core Operators
- Platform
- Gateways
The easiest way to install Kloudlite is to use our Helm charts. All the required references are available in our [Helm-chart](https://github.com/kloudlite/helm-charts) repository. The required architecture diagram is provided [here](https://kloudlite.io/docs/architecture).

## Documentation
Browse our documentation here or visit specific sections below:
- Opensource Installation: Install and run Kloudlite on your compute.
- Operators: Core operators that will run Kloudlite and its resources.
- Architecture: Kloudlite architecture diagram.
- Remote Environments: Remote environment setup and management.
- Dev Workspaces: Development environment setup and access.

## Support

Feel free to [open an issue](https://github.com/kloudlite/kloudlite/issues/new) if you have questions, encounter bugs, or have feature requests.

[Join our Discord](https://discord.gg/9FJZPHsJ) to provide feedback on in-progress features and chat with the community using Kloudlite!



