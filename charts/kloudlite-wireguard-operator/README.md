# kloudlite-wireguard-operator

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

A Helm chart for Kubernetes

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| clusterInternalDNS | string | `"cluster.local"` |  |
| image.pullPolicy | string | `"Always"` |  |
| image.repository | string | `"ghcr.io/kloudlite/kloudlite/operator/wireguard"` |  |
| image.tag | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podCIDR | string | `"10.42.0.0/16"` |  |
| publicDNSHost | string | `""` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.nameSuffix | string | `"sa"` |  |
| svcCIDR | string | `"10.43.0.0/16"` |  |
| tolerations | list | `[]` |  |

