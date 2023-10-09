apiVersion: v1
kind: Namespace
metadata:
  name: kloudlite-agent
  labels:
    kloudlite.io/created-by: kloudlite-iac

---
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: agent
  namespace: kloudlite-agent
  labels:
    kloudlite.io/created-by: kloudlite-iac
spec:
  chartRepo:
    name: kloudlite
    url: https://kloudlite.github.io/helm-charts
  chartName: kloudlite/kloudlite-agent

  chartVersion: ${kloudlite_release}

  valuesYaml: |+
    imagePullPolicy: Always
    
    # -- (string ‼️ Required) kloudlite account name
    accountName: ${kloudlite_account_name}
    
    # -- (string ‼️ Required) kloudlite cluster name
    clusterName: ${kloudlite_cluster_name}
    
    # -- (string ‼️ Required) kloudlite issued cluster token
    clusterToken: ${kloudlite_cluster_token}
    
    # -- (string) kloudlite issued access token (if already have)
    #accessToken: {{$klo}}
    
    # -- (string) cluster identity secret name, which keeps cluster token and access token
    clusterIdentitySecretName: kl-cluster-identity
    
    # -- kloudlite message office api grpc address, should be in the form of 'grpc-host:grcp-port', grpc-api.domain.com:443
    messageOfficeGRPCAddr: ${kloudlite_message_office_grpc_addr}
    #  message-office-api.dev.kloudlite.io:443
    
    # -- k8s service account name, which all the pods installed by this chart uses, will always be of format <.Release.Name>-<.Values.svcAccountName>
    svcAccountName: sa
    
    agent:
      # -- enable/disable kloudlite agent
      enabled: true
      # -- workload name for kloudlite agent
      # @ignored
      name: kl-agent
      # -- kloudlite agent image name and tag
      image: ghcr.io/kloudlite/agents/kl-agent:${kloudlite_release}
    
    # -- configuration for different kloudlite operators used in this chart
    operators:
      resourceWatcher:
        # -- enable/disable kloudlite resource watcher
        enabled: true
        # -- workload name for kloudlite resource watcher
        # @ignored
        name: kl-resource-watcher
        # -- kloudlite resource watcher image name and tag
        image: ghcr.io/kloudlite/agents/resource-watcher:${kloudlite_release}
      
      wgOperator:
        # -- whether to enable wg operator
        enabled: true
        # -- wg operator workload name
        # @ignored
        name: kl-wg-operator
        # -- wg operator image and tag
        image: ghcr.io/kloudlite/operators/wireguard:${kloudlite_release}
        
        # -- wireguard configuration options
        configuration:
          # -- cluster pods CIDR range
          podCIDR: 10.42.0.0/16
          # -- cluster services CIDR range
          svcCIDR: 10.43.0.0/16
          # -- dns hosted zone, i.e., dns pointing to this cluster, like 'clusters.kloudlite.io'
          dnsHostedZone: ${kloudlite_dns_host}
          
          # @ignored
          # -- enabled example wireguard server, and device
          enableExamples: true
    
    helmCharts:
      ingress-nginx:
        enabled: true
        name: "ingress-nginx"
        controllerKind: DaemonSet
        ingressClassName: nginx
      
      cert-manager:
        enabled: true
        name: "cert-manager"
        nodeSelector: { }
        tolerations: [ ]
        affinity: { }
      
      vector:
        enabled: true
        name: "vector"
        debugOnStdout: false