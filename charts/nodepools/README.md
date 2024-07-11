# nodepools

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

Kloudlite Nodepools enables nodepool management with kloudlite orchesterated kubernetes clusters

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| accountName | string | `""` | required only for labelling cloudprovider VMs |
| cloudprovider.k3s.joinToken | string | `""` | k3s worker nodes join token |
| cloudprovider.k3s.serverPublicHost | string | `""` | k3s masters public dns host, so that workers can join them |
| cloudprovider.name | string | `""` |  |
| cloudprovider.region | string | `""` |  |
| clusterName | string | `""` | required only for labelling cloudprovider VMs |
| kloudliteRelease | string | `""` |  |
| nodepoolJob.image.pullPolicy | string | `""` | image pull policy for kloudlite iac job, default is `Values.imagePullPolicy` |
| nodepoolJob.image.repository | string | `"ghcr.io/kloudlite/kloudlite/iac-job"` | kloudlite iac job image repository |
| nodepoolJob.image.tag | string | `""` | image tag for kloudlite iac job, by default uses `.Values.kloudliteRelease` |
| nodepoolJob.nodeAffinity | object | `{}` |  |
| nodepoolJob.nodeSelector | object | `{}` |  |
| nodepoolJob.resources.limits.cpu | string | `"500m"` |  |
| nodepoolJob.resources.limits.memory | string | `"500Mi"` |  |
| nodepoolJob.resources.requests.cpu | string | `"300m"` |  |
| nodepoolJob.resources.requests.memory | string | `"500Mi"` |  |
| nodepoolJob.tolerations | list | `[]` |  |
| nodepoolOperator.image.pullPolicy | string | `""` | image pull policy for kloudlite agent, default is `Values.imagePullPolicy` |
| nodepoolOperator.image.repository | string | `"ghcr.io/kloudlite/kloudlite/operator/nodepool"` | kloudlite agent image repository |
| nodepoolOperator.image.tag | string | `""` | image tag for kloudlite agent, by default uses `.Values.kloudliteRelease` |
| nodepoolOperator.nodeAffinity | object | `{}` |  |
| nodepoolOperator.nodeSelector | object | `{}` |  |
| nodepoolOperator.resources.limits.cpu | string | `"200m"` |  |
| nodepoolOperator.resources.limits.memory | string | `"200Mi"` |  |
| nodepoolOperator.resources.requests.cpu | string | `"100m"` |  |
| nodepoolOperator.resources.requests.memory | string | `"100Mi"` |  |
| nodepoolOperator.tolerations | list | `[]` |  |

