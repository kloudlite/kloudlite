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

    vmsingle:
      enabled: false

    vmcluster:
      enabled: true
      spec:
        vmstorage:
          storage:
            volumeClaimTemplate:
              spec:
                storageClassName: "{{.Values.persistence.storageClasses.ext4}}"
                resources:
                  requests:
                    storage: {{.Values.victoriaMetrics.configuration.vmcluster.volumeSize}}

        vmselect:
          storage:
            volumeClaimTemplate:
              spec:
                storageClassName: "{{.Values.persistence.storageClasses.ext4}}"
                resources:
                  requests:
                    storage: {{.Values.victoriaMetrics.configuration.vmselect.volumeSize}}


    nameOverride: {{.Values.victoriaMetrics.name}}
    fullnameOverride: {{.Values.victoriaMetrics.name}}
{{- end -}}
