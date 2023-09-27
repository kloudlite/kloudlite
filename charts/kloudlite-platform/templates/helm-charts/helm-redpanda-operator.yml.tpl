{{- $chartOpts := index .Values.helmCharts "redpanda-operator" }} 
{{- if $chartOpts.enabled }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: redpanda
    url: https://charts.vectorized.io

  chartName: redpanda/redpanda-operator
  chartVersion: 22.1.6

  valuesYaml: |
    nameOverride: {{$chartOpts.name}}
    fullnameOverride: {{$chartOpts.name}}

    resources: {{$chartOpts.configuration.resources}}

    webhook:
      enabled: false

{{- end }}
