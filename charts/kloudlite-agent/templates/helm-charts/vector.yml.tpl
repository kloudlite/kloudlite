{{- $chartOpts := index .Values.helmCharts "vector" }} 
{{- if $chartOpts.enabled }}

{{- $vectorSvcAccount := "vector-svc-account" }} 

{{/* INFO: Vector Svc Account is required, as we are running kubelet-metrics-reexporter as a sidecar in vector pod. This sidecar needs to access kubelet metrics and hence we need to create a service account with required permissions. */}}

apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$vectorSvcAccount}}
  namespace: {{.Release.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ $vectorSvcAccount }}-role
  namespace: {{ .Release.Namespace }}
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  - nodes
  - pods
  verbs:
  - list
  - watch

- apiGroups:
  - ""
  resources:
  - nodes/proxy
  verbs:
  - get

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{$vectorSvcAccount}}-rb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{$vectorSvcAccount}}-role
subjects:
  - kind: ServiceAccount
    name: {{$vectorSvcAccount}}
    namespace: {{.Release.Namespace}}

---

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: vector
    url: https://helm.vector.dev

  chartName: vector/vector
  chartVersion: 0.23.0
  
  jobVars:
    backOffLimit: 1
    tolerations: {{ $chartOpts.tolerations | default .Values.defaults.tolerations | toJson }}
    nodeSelector: {{ $chartOpts.nodeSelector | default .Values.defaults.nodeSelector | toJson }}

  valuesYaml: |+
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
      name: {{$vectorSvcAccount}}
    
    {{- /* WARN: specifying it is useless, but it causes helm to throw error */}}
    {{- /* refer here: https://github.com/vectordotdev/helm-charts/blob/781b414d1929826ae388e087b8d0e664fa6925b4/charts/vector/templates/NOTES.txt#L9 */}}
    quiet: true

    customConfig:
      data_dir: /vector-data-dir
      api:
        enabled: true
        address: 127.0.0.1:8686
        playground: false
      sources:
        {{- /* host_metrics: */}}
        {{- /* internal_metrics: */}}
        kubernetes_logs:
          type: kubernetes_logs
          {{/* glob_minimum_cooldown_ms: 60000 */}}
          glob_minimum_cooldown_ms: 500
        kubelet_metrics_exporter:
          type: prometheus_scrape
          endpoints:
            - http://localhost:9999/metrics/resource

      sinks:
        {{- if not $chartOpts.debugOnStdout }}
        console:
          type: console
          inputs:
            - "*"
          encoding:
            codec: json
        {{- end }}

        # -- custom configuration
        kloudlite_hosted_vector:
          type: vector
          inputs:
            - kubernetes_logs
            - kubelet_metrics_exporter
          address: {{.Values.agent.name}}.{{.Release.Namespace}}.svc.cluster.local:6000

{{- end }}
