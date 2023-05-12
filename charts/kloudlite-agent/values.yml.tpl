imagePullPolicy: Always

accountName: &accountName {{.AccountName}}
clusterName: &clusterName {{.ClusterName}}

clusterToken: {{.ClusterToken}}
accessToken: {{.AccessToken}}
clusterIdentitySecretName: {{.ClusterIdentitySecretName}}

messageOfficeGRPCAddr: &messageOfficeGRPCAddr {{.MessageOfficeGRPCAddr}}

svcAccountName: &svcAccountName {{.ClusterSvcAccountName}}

agent:
  enabled: true
  name: kl-agent
  image: {{.ImageRegistryHost}}/kloudlite/{{.EnvName}}/kl-agent:{{.ImageTag}}

operators:
  resourceWatcher:
    enabled: true
    name: kl-resource-watcher
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/resource-watcher:{{.ImageTag}}
