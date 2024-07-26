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
    url: https://charts.redpanda.com

  chartName: redpanda/operator
  {{- /* chartVersion: 5.5.1 */}}
  chartVersion: 0.4.6

  valuesYaml: |+
    nameOverride: {{$chartOpts.name}}
    fullnameOverride: {{$chartOpts.name}}

    image:
      repository: docker.redpanda.com/redpandadata/redpanda-operator
      tag: v23.2.9

  {{- /* chartName: redpanda/redpanda-operator */}}
  {{- /* chartVersion: 22.1.6 */}}
  {{- /**/}}
  {{- /* valuesYaml: | */}}
  {{- /*   nameOverride: {{$chartOpts.name}} */}}
  {{- /*   fullnameOverride: {{$chartOpts.name}} */}}
  {{- /**/}}
  {{- /*   resources: {{$chartOpts.configuration.resources}} */}}
  {{- /**/}}
  {{- /*   webhook: */}}
  {{- /*     enabled: false */}}

{{- end }}
