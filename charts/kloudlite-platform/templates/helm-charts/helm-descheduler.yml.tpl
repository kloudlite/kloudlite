{{- $chartOpts := index .Values.helmCharts "descheduler" }} 
{{- if $chartOpts.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    url: https://kubernetes-sigs.github.io/descheduler/
    name: descheduler

  chartName: "descheduler/descheduler"
  chartVersion: "0.28.0"

  valuesYaml: ""
{{- end }}
