# -- container image pull policy
imagePullPolicy: Always

# -- container image pull policy
svcAccountName: {{.SvcAccountName}}

# -- (object) node selectors for all pods in this chart
nodeSelector: {}

# -- (array) tolerations for all pods in this chart
tolerations: []

# -- (object) pod labels for all pods in this chart
podLabels: {}

# -- affine operator pods to master nodes
preferOperatorsOnMasterNodes: {{.PreferOperatorsOnMasterNodes}}

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
    enabled: false
    # -- csi drivers operator workload name
    name: kl-csi-drivers
    # -- csi drivers operator image and tag
    image: {{.ImageCsiDriversOperator}}

  routers:
    # -- whether to enable router operator
    enabled: true
    # -- router operator workload name
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
    enabled: true
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

  helmOperator:
    # -- whether to enable helm operator
    enabled: true
    # -- helm operator workload name
    name: kl-helm-operator
    # -- helm operator image and tag
    image: {{.ImageHelmOperator}}

  helmChartsOperator:
    # -- whether to enable helm-charts operator
    enabled: true
    # -- helm-charts operator workload name
    name: kl-helm-charts-operator
    # -- helm-charts operator image and tag
    image: {{.ImageHelmChartsOperator}}
