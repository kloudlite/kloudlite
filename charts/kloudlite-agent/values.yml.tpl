imagePullPolicy: Always

accountName: &accountName {{.AccountName}}
clusterName: &clusterName {{.ClusterName}}

clusterIdentitySecretName: {{.ClusterIdentitySecretName}}

messageOfficeGRPCAddr: &messageOfficeGRPCAddr {{.MessageOfficeGRPCAddr}}

svcAccountName: &svcAccountName {{.ClusterSvcAccountName}}

agent:
  enabled: true
  name: kl-agent
  image: {{.ImageRegistryHost}}/kloudlite/{{.EnvName}}/kl-agent:{{.ImageTag}}

statusAndBilling:
  enabled: true
  name: kl-status-and-billing
  image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/status-n-billing:{{.ImageTag}}
