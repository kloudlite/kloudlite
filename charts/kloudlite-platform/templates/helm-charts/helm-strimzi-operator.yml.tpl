{{- $chartOpts := index .Values.helmCharts "strimzi-operator" }} 
{{- if $chartOpts.enabled }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: strimzi
    url: https://strimzi.io/charts/

  chartName: strimzi/strimzi-kafka-operator
  chartVersion: 0.37.0

  valuesYaml: |+
    replicas: 1
    watchAnyNamespace: true
    defaultImageTag: 0.37.0

    featureGates: "+UseKRaft,+KafkaNodePools"
{{- end }}
