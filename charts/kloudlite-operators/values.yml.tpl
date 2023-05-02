imagePullPolicy: Always

accountName: "{{.AccountName}}"
region: "{{.Region}}"

clusterName: {{.ClusterName}}
svcAccountName: {{.ClusterSvcAccountName}}

nodeSelector: &nodeSelector {}
tolerations: &tolerations []

imagePullSecret:
  dockerconfigjson: {{.DockerConfigJson}}

cert-manager:
  installCRDs: true

  extraArgs:
    - "--dns01-recursive-nameservers-only"
    - "--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"

  tolerations: *tolerations
  nodeSelector: *nodeSelector

  podLabels: {}

  resources:
    limits:
      cpu: 80m
      memory: 120Mi
    requests:
      cpu: 40m
      memory: 120Mi

  webhook:
    podLabels: {}
    resources:
      limits:
        cpu: 60m
        memory: 60Mi
      requests:
        cpu: 30m
        memory: 60Mi

  cainjector:
    podLabels: {}
    resources:
      limits:
        cpu: 120m
        memory: 200Mi
      requests:
        cpu: 80m
        memory: 200Mi

wg:
  nameserverEndpoint: {{.DnsApiEndpoint}}
  nameserverUser: {{.DnsApiBasicAuthUsername}}
  nameserverPassword: {{.DnsApiBasicAuthPassword}}

  wgDomain: {{.WgDomain}}
  podCidr: {{.WgPodCIDR}}
  svcCidr: {{.WgSvcCIDR}}

agent:
  enabled: true
  name: kl-agent
  image: {{.ImageRegistryHost}}/kloudlite/{{.EnvName}}/kl-agent:{{.ImageTag}}

operators:
  project:
    enabled: true
    name: kl-projects
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/project:{{.ImageTag}}

  appAndLambda: 
    enabled: true
    name: kl-app-n-lambda
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/app-n-lambda:{{.ImageTag}}
  
  artifactsHarbor:
    enabled: true
    name: kl-artifacts-harbor
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/artifacts-harbor:{{.ImageTag}}
      
  csiDrivers:
    enabled: true
    name: kl-csi-drivers
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/csi-drivers:{{.ImageTag}}

  routers:
    enabled: true
    name: kl-routers
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/routers:{{.ImageTag}}

  msvcNMres:
    enabled: true
    name: kl-msvc-n-mres
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/msvc-n-mres:{{.ImageTag}}

  msvcMongo:
    enabled: true
    name: kl-msvc-mongo
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/msvc-mongo:{{.ImageTag}}

  msvcRedis:
    enabled: true
    name: kl-msvc-redis
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/msvc-redis:{{.ImageTag}}

  msvcRedpanda:
    enabled: true
    name: kl-redpanda
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/msvc-redpanda:{{.ImageTag}}

  msvcElasticsearch:
    enabled: false
    name: kl-msvc-elasticsearch
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/msvc-elasticsearch:{{.ImageTag}}

  {{/* statusAndBilling: */}}
  {{/*   enabled: true */}}
  {{/*   name: kl-status-n-billing */}}
  {{/*   image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/status-n-billing:{{.ImageTag}} */}}

  wgOperator:
    enabled: true
    name: kl-wg-operator
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/wg-operator:{{.ImageTag}}

  {{/* byocClientOperator: */}}
  {{/*   enabled: true */}}
  {{/*   name: kl-byoc-client-operator */}}
  {{/*   image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/byoc-client-operator:{{.ImageTag}} */}}
  
  byocOperator:
    enabled: true
    name: kl-byoc-operator
    image: {{.ImageRegistryHost}}/kloudlite/operators/{{.EnvName}}/byoc-operator:{{.ImageTag}}

  helmOperator:
    enabled: true
    name: kl-helm-operator
    image: {{.ImageRegistryHost}}/kloudlite/{{.EnvName}}/kloudlite-helm-operator:{{.ImageTag}}
