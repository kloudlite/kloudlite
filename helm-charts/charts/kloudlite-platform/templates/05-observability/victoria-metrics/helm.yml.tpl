{{- if .Values.victoriaMetrics.install }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{include "victoria-metrics.name" .}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://victoriametrics.github.io/helm-charts/
  chartName: victoria-metrics-k8s-stack

  chartVersion: {{ include "victoria-metrics.chart.version" . | quote }}

  jobVars:
    tolerations: {{.Values.scheduling.stateful.tolerations | toYaml | nindent 6}}
    nodeSelector: {{.Values.scheduling.stateful.nodeSelector | toYaml | nindent 6}}

  preInstall: |+
    curl -L0 'https://raw.githubusercontent.com/VictoriaMetrics/helm-charts/victoria-metrics-k8s-stack-{{- include "victoria-metrics.chart.version" . }}/charts/victoria-metrics-k8s-stack/charts/crds/crds/crd.yaml' > /tmp/crds.yml
    kubectl apply -f /tmp/crds.yml
    kubectl get crds
    echo "CRDs applied successfully"

  postInstall: |-
    cat > /tmp/vm-scrape.yml <<EOF
    {{- include "vector-vm-scrape" . | nindent 4 }}
    EOF

    kubectl apply -f /tmp/vm-scrape.yml
    echo "VMScraper installed"

  values:
    fullnameOverride: {{.Values.victoriaMetrics.name}}
    grafana:
      enabled: false

    crds:
      enabled: false

    victoria-metrics-operator:
      createCRD: false
      tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 8 }}
      nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 8 }}

    kube-state-metrics:
      enabled: true
      tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 8 }}
      nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 8 }}

    prometheus-node-exporter:
      enabled: true
      tolerations:
        - operator: Exists

    vmsingle:
      enabled: false

    vmagent:
      enabled: true
      spec:
        tolerations: {{.Values.scheduling.stateful.tolerations | toYaml | nindent 10 }}
        nodeSelector: {{.Values.scheduling.stateful.nodeSelector | toYaml | nindent 10 }}

    vmalert:
      enabled: true
      spec:
        tolerations: {{.Values.scheduling.stateful.tolerations | toYaml | nindent 10 }}
        nodeSelector: {{.Values.scheduling.stateful.nodeSelector | toYaml | nindent 10 }}

    alertmanager:
      enabled: true
      spec:
        tolerations: {{.Values.scheduling.stateful.tolerations | toYaml | nindent 10 }}
        nodeSelector: {{.Values.scheduling.stateful.nodeSelector | toYaml | nindent 10 }}

    vmcluster:
      enabled: true
      spec:
        retentionPeriod: {{.Values.victoriaMetrics.retentionPeriod | squote}}

        vmstorage:
          tolerations: {{.Values.scheduling.stateful.tolerations | toYaml | nindent 12 }}
          nodeSelector: {{.Values.scheduling.stateful.nodeSelector | toYaml | nindent 12 }}

          storage:
            volumeClaimTemplate:
              spec:
                storageClassName: "{{.Values.persistence.storageClasses.ext4}}"
                resources:
                  requests:
                    storage: {{.Values.victoriaMetrics.vmcluster.volumeSize}}

        vmselect:
          tolerations: {{.Values.scheduling.stateful.tolerations | toYaml | nindent 12 }}
          nodeSelector: {{.Values.scheduling.stateful.nodeSelector | toYaml | nindent 12 }}

          storage:
            volumeClaimTemplate:
              spec:
                storageClassName: "{{.Values.persistence.storageClasses.ext4}}"
                resources:
                  requests:
                    storage: {{.Values.victoriaMetrics.vmselect.volumeSize}}

        vminsert:
          tolerations: {{.Values.scheduling.stateful.tolerations | toYaml | nindent 12 }}
          nodeSelector: {{.Values.scheduling.stateful.nodeSelector | toYaml | nindent 12 }}

    nameOverride: {{ include "victoria-metrics.name" . }}
    fullnameOverride: {{ include "victoria-metrics.name" . }}
{{- end -}}
