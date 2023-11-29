apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: ${release_name}
  namespace: ${release_namespace}
  labels:
    kloudlite.io/created-by: kloudlite-iac
spec:
  chartRepo:
    name: kloudlite
    url: https://kloudlite.github.io/helm-charts

  chartName: kloudlite/kloudlite-agent
  chartVersion: ${kloudlite_release}

  jobVars:
    backOffLimit: 1
    tolerations: ${jsonencode(helm_job_tolerations)}

  valuesYaml: |+
    imagePullPolicy: Always

    accountName: ${kloudlite_account_name}
    clusterName: ${kloudlite_cluster_name}
    
    clusterToken: ${kloudlite_cluster_token}

      # -- (string) kloudlite issued access token (if already have)
    accessToken: ''

      # -- (string) cluster identity secret name, which keeps cluster token and access token
    clusterIdentitySecretName: kl-cluster-identity

      # -- kloudlite message office api grpc address, should be in the form of 'grpc-host:grcp-port', grpc-api.domain.com:443
    messageOfficeGRPCAddr: ${kloudlite_message_office_grpc_addr}

      # -- k8s service account name, which all the pods installed by this chart uses, will always be of format <.Release.Name>-<.Values.svcAccountName>
    svcAccountName: sa

      # -- cluster internal DNS, like 'cluster.local'
    clusterInternalDNS: "cluster.local"
    
    defaults:
      imageTag: ${kloudlite_release}
      imagePullPolicy: Always
      nodeSelector:
        node-role.kubernetes.io/master: "true"
      tolerations:
       - key: "node-role.kubernetes.io/master"
         operator: "Exists"
         effect: "NoSchedule"


    agent:
      # -- enable/disable kloudlite agent
      enabled: true
      # -- workload name for kloudlite agent
      # @ignored
      name: kl-agent
      # -- kloudlite agent image name and tag
      image:
        repository: ghcr.io/kloudlite/agents/kl-agent
        tag: ""
        pullPolicy: ""

    # -- (boolean) configuration for different kloudlite operators used in this chart
    preferOperatorsOnMasterNodes: true
    operators:
      resourceWatcher:
        # -- enable/disable kloudlite resource watcher
        enabled: true
        # -- workload name for kloudlite resource watcher
        # @ignored
        name: kl-resource-watcher
        # -- kloudlite resource watcher image name and tag
        image:
          repository: ghcr.io/kloudlite/operators/resource-watcher

      wgOperator:
        enabled: true
        name: kl-wg-operator
        image:
          repository: ghcr.io/kloudlite/operators/wireguard

        # -- wireguard configuration options
        configuration:
          # -- cluster pods CIDR range
          podCIDR: 10.42.0.0/16
          # -- cluster services CIDR range
          svcCIDR: 10.43.0.0/16
          # -- dns hosted zone, i.e., dns pointing to this cluster, like 'wireguard.domain.com'
          dnsHostedZone: "${kloudlite_dns_host}"

          # @ignored
          # -- enabled example wireguard server, and device
          enableExamples: false

    helmCharts:
      ingress-nginx:
        enabled: true
        name: "ingress-nginx"
        controllerKind: DaemonSet
        ingressClassName: nginx

      cert-manager:
        enabled: true
        name: "cert-manager"
        affinity: {}

      vector:
        enabled: true
        name: "vector"
        debugOnStdout: false