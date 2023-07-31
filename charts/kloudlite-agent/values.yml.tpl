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

# -- vector service account name, which all the vector pods will use
vectorSvcAccountName: &vector-svc-account-name {{.VectorSvcAccountName}}

agent:
  # -- enable/disable kloudlite agent
  enabled: true
  # -- workload name for kloudlite agent
  # @ignored
  name: kl-agent
  # -- kloudlite agent image name and tag
  image: {{.ImageAgent}}

# -- configuration for different kloudlite operators used in this chart
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
      # -- dns hosted zone, i.e. dns pointing to this cluster
      dnsHostedZone: {{.WgDnsHostedZone}}

vector:
  install: true

  role: Agent
  containerPorts:
    - containerPort: 6000

  service:
    enabled: false

  serviceHeadless:
    enabled: false

  extraContainers:
    - name: kubelet-metrics-reexporter
      image: ghcr.io/nxtcoder17/kubelet-metrics-reexporter:v1.0.0
      args:
        - --addr
        - "0.0.0.0:9999"
        {{/* - --enrich-from-labels */}}
        - --enrich-from-annotations
        - --enrich-tag
        - "kl_account_name={{.AccountName}}"
        - --enrich-tag
        - "kl_cluster_name={{.ClusterName}}"
        - --enrich-tag
        - "kl_resource_namespace={{ "{{" }}.Namespace{{ "}}" }}"
        - --filter-prefix
        - "kloudlite.io/"
        - --replace-prefix
        - "kloudlite.io/=kl_"
      env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName

  serviceAccount:
    create: false
    name: *vector-svc-account-name

  customConfig:
    data_dir: /vector-data-dir
    api:
      enabled: true
      address: 127.0.0.1:8686
      playground: false
    sources:
      host_metrics:
      internal_metrics:
      kubernetes_logs:
        type: kubernetes_logs
        {{/* glob_minimum_cooldown_ms: 60000 */}}
        glob_minimum_cooldown_ms: 500
      kubelet_metrics_exporter:
        type: prometheus_scrape
        endpoints:
          - http://localhost:9999/metrics/resource

    sinks:
      stdout: 
      {{/* stdout: */}}
      {{/*   type: console */}}
      {{/*   inputs: */}}
      {{/*     - "*" */}}
      {{/*   encoding: */}}
      {{/*     codec: json */}}

      # -- custom configuration
      kloudlite_hosted_vector:
        type: vector
        inputs:
          - kubernetes_logs
          - kubelet_metrics_exporter
        address: kl-agent.kl-init-operators.svc.cluster.local:6000

