{{- $chartOpts := .Values.kubeapps }}
{{- if $chartOpts.enabled }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://charts.bitnami.com/bitnami

  chartName: bitnami/kubeapps
  chartVersion: 14.1.2

  values: {}

{{- end }}

