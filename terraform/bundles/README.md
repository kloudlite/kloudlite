## Bundles

**Bundles** make use of different modules to create a complete infrastructure. 
They are used to create a complete infrastructure for a specific use case.

Each bundle is supposed to be composed of different modules, and is not supposed to be used as a module by any other bundle.

For example, the **[aws-k3s-HA](./aws-k3s-HA)** bundle creates a complete k3s HA cluster on AWS with all necessary components, including
- Security Groups
- EC2 Instances
- Spot Fleets
- Cloudflare DNS Records