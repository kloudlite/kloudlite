<div align="center">
  <a href="https://kloudlite.io">
    <img src="https://github.com/kloudlite/kloudlite/assets/1580519/a31a5f78-2bde-45f1-8141-d23ee8231eb1" style="height:38px" />
  </a>
  <h1>
    Development Environment as a Service  
  </h1>
  <div>No more waiting for Build & Deploy. Just code & run!</div>
  <br />
  
[Quickstart](https://kloudlite.io) | [Docs](https://kloudlite.io/docs) | [Why Kloudlite?](https://kloudlite.io/why) | [Install](https://kloudlite.io/install)

  <br />

[![discord](https://img.shields.io/discord/934762910717194260?label=discord)](https://discord.gg/m5tYzQfcG8)

[![License](https://img.shields.io/badge/License-Apache--2.0-blue)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/kloudlite/api)](https://goreportcard.com/report/github.com/kloudlite/api)
[![Issues - daytona](https://img.shields.io/github/issues/kloudlite/kloudlite)](https://github.com/kloudlite/kloudlite/issues)
![GitHub Release](https://img.shields.io/github/v/release/kloudlite/kloudlite)

[![Open Bounties](https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fkloudlite%2Fbounties%3Fstatus%3Dopen)](https://console.algora.io/org/kloudlite/bounties?status=open)
[![Rewarded Bounties](https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fkloudlite%2Fbounties%3Fstatus%3Dcompleted)](https://console.algora.io/org/kloudlite/bounties?status=completed)
</div>




Kloudlite is a platform designed to enhance developers' productivity by providing seamless, secure, and **production-parity development environments**. It connects local systems and remote environments via a WireGuard network, allowing developers to build, test, and deploy distributed applications efficiently. Kloudlite eliminates the need for separate configurations by syncing configurations and secrets across environments, and it supports collaborative coding, real-time testing, and debugging.

## Quickstart
Quickest way to start using kloudlite is to use our hosted solution. Login and setup your kloudlite account.

#### Attach Cluster
Attach your cluster in infrastructure section of kloudlite dashboard.

![Frame 2](https://github.com/kloudlite/kloudlite/assets/1580519/e9629f43-0d44-4311-b6f4-e02265ec7d3b)

#### Setup Environment
Start creating your environment. Add apps, configs, secrets in your environment.

![Frame 4](https://github.com/kloudlite/kloudlite/assets/1580519/9bb3c9ec-5c25-4a99-b038-9f17b0c40710)


#### Install Kloudlite Cli
```bash
# setup docker in your machine

# install kloudlite cli.
curl 'https://kl.kloudlite.io/kloudlite/kl!?select=kl' | bash
```

#### Access Environment
```bash

cd workspace

# setup workspace
kl init

# start development container
kl box ssh
```

## Opensource Installation
There are 3 components of kloudlite which need to be installed to run.
1. Core Operators
2. Platform
3. Wireguard Gateways

The easiest way to install kloudlite is to use our helm charts.

You can choose the kubernetes

## Documentation
Browse our docs here or visit a specific section below:
- Opensource installation: Install and run kloudlite on your compute.
- Operators: Core operators that will run kloudlite and it's resoruces.
- Architecture: Kloudlite architecture diagram.
- Remote Envs: Remote Environment setup and management
- Dev Envs: Development Environment setup and access



## Support
Feel free to open an issue if you have questions, run into bugs, or have a feature request.

Join our Discord to provide feedback on in-progress features and chat with the community using Kloudlite!

## Contributing
We are always happy to see new contributors to Kloudlite. If you are new to the Kloudlite codebase, we have a guide on how to get started. We also conduct webinar sessions every alternate week explaning the codebase and future plans. We'd love to see your contributions!

## Hiring
Apply [here](https://wellfound.com/company/kloudlite1729/) if you're interested in joining our team.
