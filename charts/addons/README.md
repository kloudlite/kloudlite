# addons

![Version: v1.0.4](https://img.shields.io/badge/Version-v1.0.4-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0.4](https://img.shields.io/badge/AppVersion-v1.0.4-informational?style=flat-square)

A Helm chart for kloudlite k3s cluster addons

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| aws.ebs_csi_driver.enabled | bool | `true` |  |
| aws.spot_node_terminator.configuration.chartVersion | string | `""` |  |
| aws.spot_node_terminator.enabled | bool | `true` |  |
| cloudprovider | string | `"aws"` | cloudprovider, should be one of the supported ones [aws] |
| common.certManager.configuration.clusterIssuers[0].acme.email | string | `"support@kloudlite.io"` |  |
| common.certManager.configuration.clusterIssuers[0].acme.server | string | `"https://acme-v02.api.letsencrypt.org/directory"` |  |
| common.certManager.configuration.clusterIssuers[0].default | bool | `true` |  |
| common.certManager.configuration.clusterIssuers[0].name | string | `"letsencrypt-prod"` |  |
| common.certManager.configuration.defaultClusterIssuer | string | `"letsencrypt-prod"` |  |
| common.certManager.configuration.nodeSelector | object | `{}` |  |
| common.certManager.configuration.tolerations | list | `[]` |  |
| common.certManager.enabled | bool | `false` |  |
| common.clusterAutoscaler.configuration.chartVersion | string | `""` |  |
| common.clusterAutoscaler.configuration.scaleDownUnneededTime | string | `"1m"` | time in golang time.Duration format like `1m or 5m`  |
| common.clusterAutoscaler.description | string | `"cluster autoscaler is useful for autoscaling nodepools in a cluster"` |  |
| common.clusterAutoscaler.enabled | bool | `true` |  |
| common.velero.configuration.backupStorage.bucket | string | `""` |  |
| common.velero.configuration.backupStorage.path | string | `""` |  |
| common.velero.configuration.backupStorage.region | string | `""` |  |
| common.velero.configuration.backupStorage.s3Url | string | `""` |  |
| common.velero.configuration.useS3Credentials.creds | object | `{"accessKey":"","secretKey":""}` | required when s3Provider is not 'aws' or pods, are not configured with Aws IAM Instance Profile |
| common.velero.configuration.useS3Credentials.enabled | string | `"true"` | if not enabled, fallsback on IAM instance profile |
| common.velero.description | string | `"velero is useful for cluster backup and restore"` |  |
| common.velero.enabled | bool | `false` |  |
| gcp.csi_driver.enabled | bool | `true` |  |
| gcp.gcloudServiceAccountCreds.json | string | `""` | base64 encoded gcp service account json |
| gcp.gcloudServiceAccountCreds.nameSuffix | string | `"gcp-creds"` |  |
| gcp.spot_node_terminator.configuration.image.repository | string | `"ghcr.io/kloudlite/kloudlite/infrastructure-as-code/gcp-spot-node-terminator"` |  |
| gcp.spot_node_terminator.configuration.image.tag | string | `""` |  |
| gcp.spot_node_terminator.enabled | bool | `true` |  |
| kloudliteRelease | string | `""` |  |
| podLabels | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `"addons-sa"` |  |

