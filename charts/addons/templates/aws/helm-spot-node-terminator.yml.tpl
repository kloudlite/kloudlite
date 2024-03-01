{{- if (eq .Values.cloudprovider "aws") }}
{{- if .Values.aws.spot_node_terminator }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: spot-node-termination-handler
  namespace: {{.Release.Namespace}}
spec:
  chartName: aws-spot-termination-handler
  chartRepoURL: https://kloudlite.github.io/helm-charts
  chartVersion: {{ .Values.aws.spot_node_terminator.configuration.chartVersion | default .Chart.AppVersion }}
  jobVars:
    tolerations:
    - operator: Exists
  {{- /* releaseName: kl-aws-spot-termination-handler */}}
  values:
    nodeSelector:
      kloudlite.io/node.is-spot: "true"
    tolerations:
      - operator: Exists
{{- end }}
{{- end }}
