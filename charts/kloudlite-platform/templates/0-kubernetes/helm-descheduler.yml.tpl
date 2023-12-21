{{- $chartOpts := index .Values.descheduler }}
{{- if $chartOpts.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: descheduler
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    url: https://kubernetes-sigs.github.io/descheduler/
    name: descheduler
  chartName: "descheduler/descheduler"
  chartVersion: "0.28.0"
  values: {}
{{- end }}
