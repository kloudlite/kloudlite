---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${svc_account_name}
  namespace: ${svc_account_namespace}

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${svc_account_name}-rb
subjects:
  - kind: ServiceAccount
    name: ${svc_account_name}
    namespace: ${svc_account_namespace}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin

---

apiVersion: batch/v1
kind: Job
metadata:
  name: install-chart-kloudlite-agent
  namespace: ${release_namespace}
spec:
  template:
    metadata:
      annotations:
        kloudlite.io/job_name: "install-chart-kloudlite-agent"
        kloudlite.io/job_type: "infrastructure-as-code"
    spec:
      restartPolicy: Never
      serviceAccountName: ${svc_account_name}
      tolerations:
        - operator: Exists
      containers:
        - name: main
          image: ghcr.io/kloudlite/job-runners/helm:${kloudlite_release}
          command:
            - bash
            - -c
            - |+
              cat > values.yml <<EOF
              # -- container image pull policy
              imagePullPolicy: Always

              # -- (string) kloudlite account name
              accountName: ${kloudlite_account_name}

              # -- (strin) kloudlite cluster name
              clusterName: ${kloudlite_cluster_name}

              # -- (string) kloudlite issued cluster token
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
                imagePullPolicy: "Always"
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
                  repository:  ghcr.io/kloudlite/api/tenant-agent
                  tag: ""
                  pullPolicy: ""
                nodeSelector: {}
                tolerations: []

              # -- (boolean) configuration for different kloudlite operators used in this chart
              preferOperatorsOnMasterNodes: true
              operators:
                agentOperator:
                  # -- enable/disable kloudlite agent operator
                  enabled: true
                  # -- workload name for kloudlite agent operator
                  name: kl-agent-operator
                  # -- kloudlite resource watcher image name and tag
                  image:
                    repository: ghcr.io/kloudlite/operator/agent
                    tag: ""
                    pullPolicy: ""
                  tolerations: []
                  nodeSelector: {}

                  configuration:
                    cloudProvider:
                      # could be aws, do, azure etc.
                      name: ${cloudprovider_name}
                      # -- cloud provider region
                      region: ${cloudprovider_region}
                    k3sJoinToken: "${k3s_agent_join_token}"
                    k3sServerPublicHost: ${k3s_masters_public_host}
                    letsEncryptSupportEmail: "support@kloudlite.io"

                wgOperator:
                  # -- whether to enable wg operator
                  enabled: true
                  # -- wg operator workload name
                  # @ignored
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
                  tolerations: []
                  nodeSelector: {}

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
                  nodeSelector: {}
                  tolerations: []
              EOF
              
              helm repo add kloudlite https://kloudlite.github.io/helm-charts
              helm repo update kloudlite
              helm upgrade --install kloudlite-agent kloudlite/kloudlite-agent --version ${kloudlite_release} --values values.yml
  backoffLimit: 0
