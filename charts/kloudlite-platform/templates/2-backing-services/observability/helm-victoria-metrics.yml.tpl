{{- if .Values.victoriaMetrics.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{.Values.victoriaMetrics.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://victoriametrics.github.io/helm-charts/
  chartName: victoria-metrics-k8s-stack

  chartVersion: 0.18.11

  values:
    fullnameOverride: {{.Values.victoriaMetrics.name}}
    grafana:
      enabled: false

    kube-state-metrics:
      enabled: true

    prometheus-node-exporter:
      enabled: true

    vmcluster:
      spec:
        vmstorage:
          storage:
            volumeClaimTemplate:
              spec:
                resources:
                  requests:
                    storage: {{.Values.victoriaMetrics.configuration.volumeSize}}

    nameOverride: {{.Values.victoriaMetrics.name}}
    fullnameOverride: {{.Values.victoriaMetrics.name}}
{{- end -}}
