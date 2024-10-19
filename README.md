<div align="center">
  <a href="https://kloudlite.io">
    <img src="https://github.com/kloudlite/kloudlite/assets/1580519/a31a5f78-2bde-45f1-8141-d23ee8231eb1" style="height:38px" />
  </a>
  <h1>
    Development Environments & Workspaces
  </h1>
  

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

**Kloudlite** is an open-source platform designed to provide seamless and secure development environments for building distributed applications. It connects local workspaces with remote Kubernetes environments via a WireGuard network, allowing developers to access services and resources with production-level parity. With Kloudlite, there’s no need for build or deploy steps during development— With service intercepts, your changes are reflected in real time, enhancing productivity and reducing the development loop.


## Architecture
![Frame 1000001531](https://github.com/user-attachments/assets/df03c018-786c-4679-aca5-15511b959331)


## Key Features:
- **WireGuard Network Integration:** Connects the workspace to environments and services using WireGuard.
- **Synchronized Workspaces:** Keeps workspace configurations and secrets in sync with connected environments and services.
- **Nix-based Package Management:** Utilizes Nix for managing workspace packages.
- **Stateless Environments:** Supports ephemeral environments without overhead.
- **Concurrent Development Support:** Enables multiple developers to work on the same environment simultaneously.
- **Application Intercepts:** Allows developers to intercept applications running in environments, redirecting their network traffic to the workspace

## Resources
- [Website](https://kloudlite.io)
- [Documentation](https://kloudlite.io/docs)
- [Roadmap](https://github.com/orgs/kloudlite/projects/22/views/5)
- [Helm Charts](https://github.com/kloudlite/helm-charts)

## Contact
- [Twitter](https://x.com/kloudlite): Follow us on Twitter!
- [Discord](https://discord.com/invite/m5tYzQfcG8): Click [here](https://discord.com/invite/m5tYzQfcG8) to join. You can ask question to our maintainers and to the rich and active community.
- [Write to us](https://kloudlite.io/contact-us)

## Security

### Reporting Security Vulnerabilities
If you've found a vulnerability or a potential vulnerability in the Kloudlite server, please let us know at support@kloudlite.io.

## License
Unless otherwise noted, the Kloudlite source files are distributed under the AGPL Version 3.0 license found in the LICENSE file.

## Shoutout to Our Open-Source Heroes 🚀

At **Kloudlite**, the open-source community is the lifeblood of our platform. We want to give a huge thanks to the following projects that form 
the foundation of Kloudlite:

- **[Kubernetes](https://github.com/kubernetes/kubernetes)**: The backbone of our environment management, enabling us to orchestrate clusters with ease and reliability.
- **[K3S](https://github.com/k3s-io/k3s)**: Lightweight and fast Kubernetes distribution that powers our local development environments.
- **[WireGuard](https://github.com/WireGuard)**: Providing secure and seamless VPN connectivity between local workspaces and remote environments.
- **[Helm](https://github.com/helm/helm)**: Simplifying our deployment process with package management for Kubernetes applications.
- **[Nix](https://github.com/nix-community)**: Managing dependencies in our development containers, ensuring flexibility and consistency across environments.
- **[Docker](https://github.com/docker)**: Containerizing our applications to provide consistency and simplicity across setups.
- **[MongoDB](https://github.com/mongodb/mongo)**: Powering our data storage with its flexible, scalable document-based architecture.
- **[NATS](https://github.com/nats-io/nats-server)**: Enabling fast and lightweight real-time messaging and communication between services.
- **[VictoriaMetrics](https://github.com/VictoriaMetrics/VictoriaMetrics)**: Handling our monitoring and metrics with high performance and scalability.
- **[Apollo Federation](https://github.com/apollographql/federation)**: Orchestrating our distributed GraphQL architecture for seamless communication across services.
- **[gqlgen](https://github.com/99designs/gqlgen)**: Powering our Golang-based GraphQL server, ensuring type safety and performance.
- **[Remix](https://github.com/remix-run)**: Providing a modern and flexible framework for building fast, dynamic, and reliable frontend experiences.
- **[TailwindCSS](https://github.com/tailwindlabs/tailwindcss)**: Simplifying our frontend design with a utility-first CSS framework for beautiful UIs.
- **[Operator Framework](https://github.com/operator-framework)**: Helping us build powerful and reliable Kubernetes operators to automate complex tasks.

We are deeply grateful to the maintainers and contributors of all these projects for driving innovation and making open-source accessible. Your work powers the heart of Kloudlite! 👏
