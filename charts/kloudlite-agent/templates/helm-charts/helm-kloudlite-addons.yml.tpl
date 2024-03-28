{{- if .Values.helmCharts.kloudliteAddons.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: "kloudlite-addons"
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://kloudlite.github.io/helm-charts
  chartName: addons
  chartVersion: {{.Values.helmCharts.kloudliteAddons.configuration.chartVersion | default (include "image-tag" .) }}

  jobVars:
    backOffLimit: 1
    nodeSelector: {{ .Values.helmCharts.kloudliteAddons.configuration.nodeSelector | default .Values.nodeSelector | toYaml | nindent 8 }}
    tolerations: {{ .Values.helmCharts.kloudliteAddons.configuration.tolerations | default .Values.tolerations | toYaml | nindent 8 }}

  values:
    cloudprovider: "{{.Values.cloudProvider}}"

    aws:
      ebs_csi_driver:
        enabled: true

      spot_node_terminator:
        enabled: true

    common:
      clusterAutoscaler:
        enabled: {{.Values.helmCharts.kloudliteAddons.configuration.clusterAutoscaler.enabled}}
        configuration:
          scaleDownUnneededTime: 3m

      velero:
        enabled: false

{{- end }}
