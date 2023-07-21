{{- $redpandaOperatorName := include "redpanda-operator.name" . }} 

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$redpandaOperatorName}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: redpanda
    url: https://charts.vectorized.io

  chartName: redpanda/redpanda-operator
  chartVersion: 22.1.6

  valuesYaml: |
    nameOverride: {{ $redpandaOperatorName }}
    fullnameOverride: {{ $redpandaOperatorName }}

    resources:
      limits:
        cpu: 60m
        memory: 60Mi
      requests:
        cpu: 40m
        memory: 40Mi

    webhook:
      enabled: false
