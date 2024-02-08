{{- $chartOpts := index .Values.helmCharts.clusterAutoscaler }}
{{- if $chartOpts.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: cluster-autoscaler
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://kloudlite.github.io/helm-charts
  chartName: "kloudlite-autoscalers"
  chartVersion: {{ include "image-tag" .}}
  jobVars:
    tolerations:
      - operator: Exists
  values:
    kloudliteRelease: {{ include "image-tag" .}}
    clusterAutoscaler:
      enabled: true
      nodeSelector:
        node-role.kubernetes.io/master: "true"
      tolerations:
        - operator: Exists
      configuration:
        scaleDownUnneededTime: 1m
{{- end }}
