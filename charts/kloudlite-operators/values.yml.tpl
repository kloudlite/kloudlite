# -- container image pull policy
imagePullPolicy: Always

# -- container image pull policy
svcAccountName: {{.ClusterSvcAccountName}}

# -- default image pull secret name
defaultImagePullSecretName: {{.DefaultImagePullSecretName}}

# -- (object) node selectors for all pods in this chart
nodeSelector: &nodeSelector {}

# -- (array) tolerations for all pods in this chart
tolerations: &tolerations []

# -- (object) pod labels for all pods in this chart
podLabels: &podLabels {}

{{- /* imagePullSecret: */ -}}
{{- /*   dockerconfigjson: {{.DockerConfigJson}} */ -}}

# -- configuration option for cert-manager (https://cert-manager.io/docs/installation/helm/)
cert-manager:
  # -- cert-manager whether to install CRDs
  installCRDs: true

  extraArgs:
    - "--dns01-recursive-nameservers-only"
    - "--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"

  tolerations: *tolerations
  nodeSelector: *nodeSelector

  podLabels: *podLabels

  resources:
    # -- resource limits for cert-manager controller pods
    limits:
      # -- cpu limit for cert-manager controller pods
      cpu: 80m
      # -- memory limit for cert-manager controller pods
      memory: 120Mi
    requests:
      # -- cpu request for cert-manager controller pods
      cpu: 40m
      # -- memory request for cert-manager controller pods
      memory: 120Mi

  webhook:
    podLabels: *podLabels
    # -- resource limits for cert-manager webhook pods
    resources:
      # -- resource limits for cert-manager webhook pods
      limits:
        # -- cpu limit for cert-manager webhook pods
        cpu: 60m
        # -- memory limit for cert-manager webhook pods
        memory: 60Mi
      requests:
        # -- cpu limit for cert-manager webhook pods
        cpu: 30m
        # -- memory limit for cert-manager webhook pods
        memory: 60Mi

  cainjector:
    podLabels: *podLabels
    # -- resource limits for cert-manager cainjector pods
    resources:
      # -- resource limits for cert-manager webhook pods
      limits:
        # -- cpu limit for cert-manager cainjector pods
        cpu: 120m
        # -- memory limit for cert-manager cainjector pods
        memory: 200Mi
      requests:
        # -- cpu requests for cert-manager cainjector pods
        cpu: 80m
        # -- memory requests for cert-manager cainjector pods
        memory: 200Mi

# -- wireguard configuration options
wg:
  # -- dns nameserver http endpoint
  nameserver:
    endpoint: {{.DnsApiEndpoint}}
    # -- basic auth configurations for dns nameserver http endpoint
    basicAuth:
      # -- whether to enable basic auth for dns nameserver http endpoint
      enabled: {{.DnsApiBasicAuthEnabled}}
      # -- if enabled, basic auth username for dns nameserver http endpoint
      username: {{.DnsApiBasicAuthUsername}}
      # -- if enabled, basic auth password for dns nameserver http endpoint
      password: {{.DnsApiBasicAuthPassword}}

  # -- baseDomain for wireguard service, to be exposed
  baseDomain: {{.WgDomain}}
  # -- cluster pods CIDR range
  podCidr: {{.WgPodCIDR}}
  # -- cluster services CIDR range
  svcCidr: {{.WgSvcCIDR}}

operators:
  project:
    # -- whether to enable project operator
    enabled: true
    # -- project operator workload name
    name: kl-projects
    # -- project operator image and tag
    image: {{.ImageProjectOperator}}

  app: 
    # -- whether to enable app operator
    enabled: true
    # -- app operator workload name
    name: kl-app
    # -- app operator image and tag
    image: {{.ImageAppOperator}}

  csiDrivers:
    # -- whether to enable csi drivers operator
    enabled: true
    # -- csi drivers operator workload name
    name: kl-csi-drivers
    # -- csi drivers operator image and tag
    image: {{.ImageCsiDriversOperator}}

  routers:
    # -- whether to enable routers operator
    enabled: true
    # -- routers operator workload name
    name: kl-routers
    # -- routers operator image and tag
    image: {{.ImageRoutersOperator}}

  msvcNMres:
    # -- whether to enable msvc-n-mres operator
    enabled: true
    # -- msvc-n-mres operator workload name
    name: kl-msvc-n-mres
    # -- msvc-n-mres operator image and tag
    image: {{.ImageMsvcNMresOperator}}

  msvcMongo:
    # -- whether to enable msvc-mongo operator
    enabled: true
    # -- msvc mongo operator workload name
    name: kl-msvc-mongo
    # -- name msvc mongo operator image and tag
    image: {{.ImageMsvcMongoOperator}}

  msvcRedis:
    # -- whether to enable msvc-redis operator
    enabled: true
    # -- msvc redis operator workload name
    name: kl-msvc-redis
    # -- msvc redis operator image and tag
    image: {{.ImageMsvcRedisOperator}}

  msvcRedpanda:
    # -- whether to enable msvc-redpanda operator
    enabled: false
    # -- msvc redpanda operator workload name
    name: kl-redpanda
    # -- msvc redpanda operator image and tag
    image: {{.ImageMsvcRedpandaOperator}}

  msvcElasticsearch:
    # -- whether to enable msvc-elasticsearch operator
    enabled: false
    # -- msvc elasticsearch operator workload name
    name: kl-msvc-elasticsearch
    # -- msvc elasticsearch operator image and tag
    image: {{.ImageMsvcElasticsearchOperator}}

  wgOperator:
    # -- whether to enable wg operator
    enabled: false
    # -- wg operator workload name
    name: kl-wg-operator
    # -- wg operator image and tag
    image: {{.ImageWgOperator}}

  helmOperator:
    # -- whether to enable helm operator
    enabled: true
    # -- helm operator workload name
    name: kl-helm-operator
    # -- helm operator image and tag
    image: {{.ImageHelmOperator}}
