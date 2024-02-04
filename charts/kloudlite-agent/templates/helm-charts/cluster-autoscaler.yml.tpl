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
    defaults:
      imagePullPolicy: {{ include "image-pull-policy" . }}

    serviceAccount:
      create: true
      nameSuffix: "sa"

    clusterAutoscaler:
      enabled: true
      image:
        repository: "ghcr.io/kloudlite/autoscaler/cluster-autoscaler"
        tag: ""
      nodeSelector:
        node-role.kubernetes.io/master: "true"
      tolerations:
        - operator: Exists
      configuration:
        scaleDownUnneededTime: 1m
{{- end }}
