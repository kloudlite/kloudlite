apiVersion: msvc.kloudlite.io/v1
kind: HelmOpenSearch
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  # Default values copied from <project_dir>/helm-charts/opensearch/values.yaml
  antiAffinity: soft
  antiAffinityTopologyKey: kubernetes.io/hostname
  clusterName: opensearch-cluster
  config:
    opensearch.yml: |
      cluster.name: opensearch-cluster
      
      # Bind to all interfaces because we don't know what IP address Docker will assign to us.
      network.host: 0.0.0.0
      
      # Setting network.host to a non-loopback address enables the annoying bootstrap checks. "Single-node" mode disables them again.
      # discovery.type: single-node
      
      # Start OpenSearch Security Demo Configuration
      # WARNING: revise all the lines below before you go into production
      plugins:
        security:
          ssl:
            transport:
              pemcert_filepath: esnode.pem
              pemkey_filepath: esnode-key.pem
              pemtrustedcas_filepath: root-ca.pem
              enforce_hostname_verification: false
            http:
              enabled: true
              pemcert_filepath: esnode.pem
              pemkey_filepath: esnode-key.pem
              pemtrustedcas_filepath: root-ca.pem
          allow_unsafe_democertificates: true
          allow_default_init_securityindex: true
          authcz:
            admin_dn:
              - CN=kirk,OU=client,O=client,L=test,C=de
          audit.type: internal_opensearch
          enable_snapshot_restore_privilege: true
          check_snapshot_restore_write_privileges: true
          restapi:
            roles_enabled: ["all_access", "security_rest_api_access"]
          system_indices:
            enabled: true
            indices:
              [
                ".opendistro-alerting-config",
                ".opendistro-alerting-alert*",
                ".opendistro-anomaly-results*",
                ".opendistro-anomaly-detector*",
                ".opendistro-anomaly-checkpoints",
                ".opendistro-anomaly-detection-state",
                ".opendistro-reports-*",
                ".opendistro-notifications-*",
                ".opendistro-notebooks",
                ".opendistro-asynchronous-search-response*",
              ]
      ######## End OpenSearch Security Demo Configuration ########
  enableServiceLinks: true
  envFrom: []
  extraContainers: []
  extraEnvs: []
  extraInitContainers: []
  extraObjects: []
  extraVolumeMounts: []
  extraVolumes: []
  fsGroup: ""
  fullnameOverride: ""
  global:
    dockerRegistry: ""
  hostAliases: []
  httpPort: 9200
  image:
    pullPolicy: IfNotPresent
    repository: opensearchproject/opensearch
    tag: ""
  imagePullSecrets: []
  ingress:
    annotations: {}
    enabled: false
    hosts:
      - chart-example.local
    path: /
    tls: []
  initResources: {}
  keystore: []
  labels: {}
  lifecycle: {}
  majorVersion: ""
  masterService: opensearch-cluster-master
  masterTerminationFix: false
  maxUnavailable: 1
  nameOverride: ""
  networkHost: 0.0.0.0
  networkPolicy:
    create: false
    http:
      enabled: false
  nodeAffinity: {}
  nodeGroup: master
  nodeSelector: {}
  opensearchHome: /usr/share/opensearch
  opensearchJavaOpts: -Xmx512M -Xms512M
  persistence:
    accessModes:
      - ReadWriteOnce
    annotations: {}
    enableInitChown: true
    enabled: true
    labels:
      enabled: false
    size: 8Gi
  plugins:
    enabled: false
    installList: []
  replicas: 3
  resources:
    requests:
      cpu: 1000m
      memory: 100Mi
  roles:
    - master
    - ingest
    - data
    - remote_cluster_client
  schedulerName: ""
  secretMounts: []
  securityConfig:
    actionGroupsSecret: null
    config:
      data: {}
      dataComplete: true
      securityConfigSecret: ""
    configSecret: null
    enabled: true
    internalUsersSecret: null
    path: /usr/share/opensearch/plugins/opensearch-security/securityconfig
    rolesMappingSecret: null
    rolesSecret: null
    tenantsSecret: null
  securityContext:
    capabilities:
      drop:
        - ALL
    runAsNonRoot: true
    runAsUser: 1000
  service:
    annotations: {}
    externalTrafficPolicy: ""
    headless:
      annotations: {}
    httpPortName: http
    labels: {}
    labelsHeadless: {}
    loadBalancerIP: ""
    loadBalancerSourceRanges: []
    nodePort: ""
    transportPortName: transport
    type: ClusterIP
  sidecarResources: {}
  sysctl:
    enabled: false
  sysctlVmMaxMapCount: 262144
  terminationGracePeriod: 120
  tolerations: []
  topologySpreadConstraints: []
  transportPort: 9300
  updateStrategy: RollingUpdate


