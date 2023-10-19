# -- container image pull policy
imagePullPolicy: Always

# -- (string {{.Required}}) kloudlite account name
accountName: {{.AccountName }}

# -- (string {{.Required}}) kloudlite cluster name
clusterName: {{.ClusterName}}

# -- (string {{.Required}}) kloudlite issued cluster token
clusterToken: {{.ClusterToken}}

# -- (string) kloudlite issued access token (if already have)
accessToken: {{ .AccessToken | default "" }}

# -- (string) cluster identity secret name, which keeps cluster token and access token
clusterIdentitySecretName: {{.ClusterIdentitySecretName}}

# -- kloudlite message office api grpc address, should be in the form of 'grpc-host:grcp-port', grpc-api.domain.com:443
messageOfficeGRPCAddr: {{.MessageOfficeGRPCAddr}}

# -- k8s service account name, which all the pods installed by this chart uses, will always be of format <.Release.Name>-<.Values.svcAccountName>
svcAccountName: {{.SvcAccountName}}

# -- cluster internal DNS, like 'cluster.local'
clusterInternalDNS: "cluster.local"

agent:
  # -- enable/disable kloudlite agent
  enabled: true
  # -- workload name for kloudlite agent
  # @ignored
  name: kl-agent
  # -- kloudlite agent image name and tag
  image: {{.ImageAgent}}

# -- (boolean) configuration for different kloudlite operators used in this chart
preferOperatorsOnMasterNodes: {{.PrefersOperatorsOnMasterNodes}}
operators:
  resourceWatcher:
    # -- enable/disable kloudlite resource watcher
    enabled: true
    # -- workload name for kloudlite resource watcher
    # @ignored
    name: kl-resource-watcher
    # -- kloudlite resource watcher image name and tag
    image: {{.ImageOperatorResourceWatcher }}

  wgOperator:
    # -- whether to enable wg operator
    enabled: true
    # -- wg operator workload name
    # @ignored
    name: kl-wg-operator
    # -- wg operator image and tag
    image: {{.ImageWgOperator}}

    # -- wireguard configuration options
    configuration:
      # -- cluster pods CIDR range
      podCIDR: {{.WgPodCIDR}}
      # -- cluster services CIDR range
      svcCIDR: {{.WgSvcCIDR}}
      # -- dns hosted zone, i.e., dns pointing to this cluster, like 'clusters.kloudlite.io'
      dnsHostedZone: {{.WgDnsHostedZone}}

      # @ignored
      # -- enabled example wireguard server, and device
      enableExamples: {{.EnableWgExamples}}

helmCharts:
  ingress-nginx:
    enabled: true
    name: "ingress-nginx"
    controllerKind: DaemonSet
    ingressClassName: nginx
    {{- /* affinity: {} */}}

  cert-manager:
    enabled: true
    name: "cert-manager"
    nodeSelector: {}
    tolerations: []
    affinity: {}

  vector:
    enabled: true
    name: "vector"
    debugOnStdout: false
