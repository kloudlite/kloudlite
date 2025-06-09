{{- $accountName := get . "account-name" }}
{{- $clusterName := get . "cluster-name" }}

{{- $kloudliteRelease := get . "kloudlite-release" }}
{{- $messageOfficeGrpcAddr := get . "message-office-grpc-addr" }}

{{- $clusterToken := get . "cluster-token" }}
{{- $publicDNSHost := get . "public-dns-host" }}

{{- $cloudprovider := get . "cloudprovider" }}

{{- $gcpServiceAccountJson := get . "gcp-service-account-json" | default "" }}

{{- if and (eq $cloudprovider "gcp") (not $gcpServiceAccountJson) }}
{{fail "when cloudprovider is gcp, field `gcp-service-account-json` is required"}}
{{- end }}

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
    kubectl apply -f https://github.com/kloudlite/helm-charts/releases/download/{{$kloudliteRelease}}/crds-kloudlite.yml --server-side

  postInstall: |+
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

    nodeSelector:
      node-role.kubernetes.io/master: "true"
    tolerations:
      - key: "node-role.kubernetes.io/master"
        operator: "Exists"
        effect: "NoSchedule"

    cloudProvider: "{{$cloudprovider}}"

    {{- if (eq $cloudprovider "gcp") }}
    gcp:
      gcloudServiceAccountCreds:
        # -- base64 encoded service account json
        json: {{$gcpServiceAccountJson}}
    {{- end }}

    agent:
      enabled: true
      name: kl-agent
      nodeSelector: {}
      tolerations: []

    preferOperatorsOnMasterNodes: true
    operators:
      agentOperator:
        enabled: true
        name: kl-agent-operator

        tolerations: []
        nodeSelector: {}

        configuration:
          routers:
            letsEncryptSupportEmail: "support@kloudlite.io"

          nodepools:
            enabled: true
            # must be one of aws,azure,gcp
            cloudprovider: "{{$cloudprovider}}"

          wireguard:
            publicDNSHost: {{$publicDNSHost}}

          {{- /* wireguard: */}}
          {{- /*   podCIDR: 10.42.0.0/16 */}}
          {{- /*   svcCIDR: 10.43.0.0/16 */}}
          {{- /**/}}
          {{- /*   deviceNamespace: kl-vpn-devices */}}

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

      clusterAutoscaler:
        enabled: true
