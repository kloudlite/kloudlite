{{- if .Values.grafana.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{.Values.grafana.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://grafana.github.io/helm-charts
  chartName: grafana
  chartVersion: 7.0.22

  values:
    replicas: 1

    nodeSelector: {{.Values.grafana.configuration.nodeSelector | toJson }}
    priorityClassName: {{.Values.global.statefulPriorityClassName}}

    persistence:
      type: statefulset
      storageClassName: {{.Values.persistence.storageClasses.ext4}}
      size: {{.Values.grafana.configuration.volumeSize}}

    resources:
      limits:
        cpu: 300m
        memory: 300Mi
      requests:
        cpu: 200m
        memory: 200Mi
{{- end }}

