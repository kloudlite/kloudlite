{{- $chartOpts := index .Values.helmCharts "grafana" }} 
{{- if $chartOpts.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: bitnami
    url: https://charts.bitnami.com/bitnami

  chartName: bitnami/grafana
  chartVersion: 9.0.1

  valuesYaml: |+
    global:
      storageClass: {{.Values.persistence.storageClasses.ext4}}

    nameOverride: {{$chartOpts.name}}
    fullnameOverride: {{$chartOpts.name}}

    persistence:
      enabled: true
      size: {{$chartOpts.configuration.volumeSize}}

{{- end }}

