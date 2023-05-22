imagePullPolicy: Always

accountName: {{.AccountName}}
clusterName: {{.ClusterName}}

clusterToken: {{.ClusterToken}}
accessToken: {{.AccessToken | default "" }}
clusterIdentitySecretName: {{.ClusterIdentitySecretName}}

defaultImagePullSecretName: {{.DefaultImagePullSecretName}}

messageOfficeGRPCAddr: {{.MessageOfficeGRPCAddr}}

svcAccountName: {{.ClusterSvcAccountName}}

agent:
  enabled: true
  name: kl-agent
  image: {{.ImageRegistryHost}}/kloudlite/{{.EnvName}}/kl-agent:{{.ImageTag}}

operators:
  resourceWatcher:
    enabled: true
    name: kl-resource-watcher
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/resource-watcher:{{.ImageTag}}
