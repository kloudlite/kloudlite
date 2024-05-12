<p>
  <img width=300 src="https://github.com/kloudlite/kloudlite/assets/1580519/27001f02-a87f-46b7-aaaf-3b36bafc73e0" alt="KloudLite Logo">
</p>

<p>
  Cloud Native RemoteLocal Environments to build distributed applications.
</p>

Kloudlite is a remote local environments platform designed to enhance productivity for developers working on distributed applications. By leveraging Kubernetes, Kloudlite provides a seamless bridge between local systems and remote environments, ensuring efficient and effective development workflows.

## Features
- **Isolated and Replicable Local Development Containers**: Kloudlite provides developers with isolated local development containers that seamlessly connect with remote environments. This setup ensures a replicable development process that is both efficient and consistent across different machines and team members.

- **Comprehensive Environment Management**: Kloudlite environments handle configurations, secrets, and access to managed services, applications, and external services efficiently. These environments are designed to meet all your development needs. Since they are stateless and ephemeral, developers can easily clone them for isolated, parallel development efforts.

- **Application/Service Interception**: With Kloudlite, developers can intercept and replace an application or service running in the remote environment with the version running in their local IDE. This feature allows for real-time testing and debugging before proceeding with builds and deployments, significantly enhancing development efficiency.

- **Collaborative Development**: Kloudlite supports collaborative development by allowing multiple developers to connect to the same environment. Team members can intercept their respective services and debug collaboratively, directly from their local machines, fostering teamwork and streamlining problem-solving.

## Installation

To install Kloudlite using Helm, run the following command in your terminal:

```bash
helm install [NAME] [CHART] [flags]
```
Replace [NAME] with the name you want to assign to your Kloudlite installation, and [CHART] with the appropriate Helm chart for Kloudlite.

## Getting Started
Follow these steps to begin using Kloudlite:
- **Add Cluster to Kloudlite:** Start by adding your Kubernetes cluster to Kloudlite. This integration is crucial for managing your environments and services.
- **Create Your First Environment:** Once your cluster is added, create your first environment. This is where you'll manage your services and applications.
- **Configure Local Development Container:** Configure your local development container by setting up the kl.yaml file. This file will dictate how your local environment connects and interacts with your remote Kloudlite environment.
- **Learn More:** To dive deeper into Kloudlite's features and capabilities, visit our documentation.

## Documentation
For more detailed information and step-by-step guides, please visit our [documentation](https://kloudlite.com/docs).

## Support
If you encounter any issues or require assistance, please visit our [support page](https://kloudlite.com/support) or reach out to our community on [Discord](https://discord.gg/kloudlite).

## Contributing
We welcome contributions from the community! If you're interested in making Kloudlite better, please refer to our [contributing guidelines](https://kloudlite.com/contribute).

## License
Kloudlite is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE.md) file for more details.


<p align="center">
  Made with ❤️ by the Kloudlite team
</p>
