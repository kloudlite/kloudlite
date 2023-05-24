# --  container image pull policy
imagePullPolicy: Always

# -- kloudlite account name
accountName: {{.AccountName }}

# --  kloudlite cluster name
clusterName: {{.ClusterName}}

# --  kloudlite issued cluster token
clusterToken: {{.ClusterToken}}

# -- kloudlite issued access token (if already have)
accessToken: {{ .AccessToken | default "" }}

# -- cluster identity secret name, which keeps cluster token and access token
clusterIdentitySecretName: {{.ClusterIdentitySecretName}}

# -- default image pull secret name, defaults to {{.DefaultImagePullSecretName}}
defaultImagePullSecretName: {{.DefaultImagePullSecretName}}

# -- kloudlite message office api grpc address, should be in the form of 'grpc-host:grcp-port'
messageOfficeGRPCAddr: {{.MessageOfficeGRPCAddr}}

# -- k8s service account name, which all the pods installed by this chart uses
svcAccountName: {{.ClusterSvcAccountName}}

agent:
  # -- enable/disable kloudlite agent
  enabled: true
  # -- workload name for kloudlite agent
  name: kl-agent
  # -- kloudlite agent image name and tag
  image: {{.ImageAgent}}

operators:
  # -- configuration for different kloudlite operators used in this chart
  resourceWatcher:
    # -- enable/disable kloudlite resource watcher
    enabled: true
    # -- workload name for kloudlite resource watcher
    name: kl-resource-watcher
    # -- kloudlite resource watcher image name and tag
    image: {{.ImageOperatorResourceWatcher }}
