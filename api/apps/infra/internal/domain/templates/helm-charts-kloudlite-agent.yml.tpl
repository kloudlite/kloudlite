{{- $accountName := get . "account-name" }}
{{- $clusterName := get . "cluster-name" }}

{{- $kloudliteRelease := get . "kloudlite-release" }}
{{- $messageOfficeGrpcAddr := get . "message-office-grpc-addr" }}

{{- $clusterToken := get . "cluster-token" }}

---
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: kloudlite-agent
  namespace: kloudlite
spec:
  chartRepoURL: https://kloudlite.github.io/helm-charts
  chartVersion: {{$kloudliteRelease}}
  chartName: kloudlite-agent

  jobVars:
    tolerations:
      - operator: Exists
    
  preInstall: |+
    if kubectl get ns kloudlite-tmp;
    then
      kubectl delete ns kloudlite-tmp
    fi
    
  values:
    imagePullPolicy: Always

    accountName: {{$accountName}}
    clusterName: {{$clusterName}}

    clusterToken: {{$clusterToken}}

    clusterIdentitySecretName: kl-cluster-identity

    messageOfficeGRPCAddr: {{$messageOfficeGrpcAddr}}

    svcAccountName: sa

    clusterInternalDNS: "cluster.local"

    defaults:
      imageTag: {{$kloudliteRelease}}
      imagePullPolicy: "Always"
      nodeSelector:
        node-role.kubernetes.io/master: "true"
      tolerations:
        - key: "node-role.kubernetes.io/master"
          operator: "Exists"
          effect: "NoSchedule"

    agent:
      enabled: true
      name: kl-agent
      image:
        repository:  ghcr.io/kloudlite/api/tenant-agent
        tag: ""
        pullPolicy: ""
      nodeSelector: {}
      tolerations: []

    preferOperatorsOnMasterNodes: true
    operators:
      agentOperator:
        enabled: true
        name: kl-agent-operator
        image:
          repository: ghcr.io/kloudlite/operator/agent
          tag: ""
          pullPolicy: ""

        tolerations: []
        nodeSelector: {}

        configuration:
          letsEncryptSupportEmail: "support@kloudlite.io"

      wgOperator:
        enabled: true
        name: kl-wg-operator
        # -- wg operator image and tag
        image:
          repository: ghcr.io/kloudlite/operator/wireguard
          tag: ""
          pullPolicy: ""

        tolerations: []
        nodeSelector: {}

        # -- wireguard configuration options
        configuration:
          # -- cluster pods CIDR range
          podCIDR: 10.42.0.0/16
          # -- cluster services CIDR range
          svcCIDR: 10.43.0.0/16
          # -- dns hosted zone, i.e., dns pointing to this cluster, like 'wireguard.domain.com'

          enableExamples: false

    helmCharts:
      ingressNginx:
        enabled: true
        name: "ingress-nginx"
        tolerations: []
        nodeSelector: {}
        configuration:
          controllerKind: DaemonSet
          ingressClassName: nginx

      certManager:
        enabled: true
        name: "cert-manager"
        nodeSelector: {}
        tolerations: []
        affinity: {}
        configuration:
          defaultClusterIssuer: letsencrypt-prod
          clusterIssuers:
            - name: letsencrypt-prod
              default: true
              acme:
                email: "support@kloudlite.io"
                server: https://acme-v02.api.letsencrypt.org/directory

      vector:
        enabled: true
        name: "vector"
        debugOnStdout: false
        nodeSelector: {}
        tolerations: []

