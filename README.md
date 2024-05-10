<p>
  <img width=300 src="https://github.com/kloudlite/kloudlite/assets/1580519/27001f02-a87f-46b7-aaaf-3b36bafc73e0" alt="KloudLite Logo">
</p>

<p>
  Cloud Native RemoteLocal Environments to build distributed applications.
</p>

## Components
### Core
- API - Backend microservices
- Web - Frontend
- kl - Cli

### Kubernetes Resource Management
- Operators - Kubernetes Operators

### Managed kubernetes
- Kloudlite Autoscaler - Cluster autoscaler
- Infrastructure As Code - Terraform modules

### Observibility
- Kubelet Metrics ReExporter - Observibility

### Distribution
- Helm Charts - Package Distribution

## Features
- **Seamless Integration:** Sync configurations, secrets, and development containers between local systems and remote environments.
- **Reduced Latency:** Work with local IDEs while seamlessly interacting with remote resources, reducing latency issues.
- **Debugging Support:** Use service interception for debugging and troubleshooting distributed applications.
- **Collaborative Coding:** Collaborate effectively with team members, accelerating the development cycle.

## Installation

### Setup on Kubernetes
You can install KloudLite platform on Kubernetes using Helm.
```
helm repo add kloudlite https://kloudlite.github.io/helm-charts
helm repo update

helm install [RELEASE_NAME] kloudlite/kloudlite-platform --namespace [NAMESPACE] [--create-namespace]
```

### CLI Local Installation
To access KloudLite remote-local environments for development, you can use the CLI. To install the CLI, execute the following command:
```
curl 'https://kl.kloudlite.io/kloudlite!?select=kl' | bash
```

## Usage
To use KloudLite, follow these steps:
1. Configure your local environment settings.
2. Connect to your remote environment using the KloudLite CLI.
3. Start developing your distributed application with reduced latency and enhanced collaboration.

## Contributing
We welcome contributions from the community! If you'd like to contribute to KloudLite, please follow these guidelines:
1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature-name`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature/your-feature-name`).
5. Create a new Pull Request.

## License
KloudLite is licensed under the Apache License. See the LICENSE file for details.


---
<p align="center">
  Made with ❤️ by the Kloudlite team
</p>
