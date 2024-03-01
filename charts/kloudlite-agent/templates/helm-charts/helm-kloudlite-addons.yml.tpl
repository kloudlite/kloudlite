{{- $chartOpts :=  .Values.helmCharts.certManager}} 
{{- if $chartOpts.enabled }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://kloudlite.github.io/helm-charts
  chartName: kloudlite-addons
  chartVersion: {{.Values.helmCharts.kloudliteAddons.configuration.chartVersion | default (include "image-tag" .) }}

  jobVars:
    backOffLimit: 1
    nodeSelector: {{ $chartOpts.nodeSelector | default .Values.nodeSelector | toYaml | nindent 8 }}
    tolerations: {{ $chartOpts.tolerations | default .Values.tolerations | toYaml | nindent 8 }}

  values:
    cloudprovider: "{{.Values.cloudProvider}}"

    aws:
      ebs_csi_driver:
        enabled: true

      spot_node_terminator:
        enabled: true

    common:
      clusterAutoscaler:
        enabled: true
        configuration:
          scaleDownUnneededTime: 3m
      velero:
        enabled: false

{{- end }}
