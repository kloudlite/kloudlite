{{- $chartOpts := index .Values.descheduler }}
{{- if $chartOpts.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: descheduler
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://kubernetes-sigs.github.io/descheduler/
  chartName: "descheduler/descheduler"
  chartVersion: "0.28.0"
  values: {}
{{- end }}
