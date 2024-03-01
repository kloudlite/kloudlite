{{- if .Values.common.clusterAutoscaler.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: cluster-autoscaler
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://kloudlite.github.io/helm-charts
  chartName: "kloudlite-autoscalers"
  chartVersion: {{ .Values.common.clusterAutoscaler.configuration.chartVersion | default .Chart.AppVersion }}
  jobVars:
    tolerations:
      - operator: Exists
  values:
    clusterAutoscaler:
      enabled: true
      nodeSelector:
        node-role.kubernetes.io/master: "true"
      tolerations:
        - operator: Exists
      configuration:
        scaleDownUnneededTime: {{.Values.common.clusterAutoscaler.configuration.scaleDownUnneededTime}}
{{- end }}
