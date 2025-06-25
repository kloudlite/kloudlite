{{- define "helmchart-operator-env" -}}
- name: HELM_JOB_RUNNER_IMAGE
  value: {{.Values.agentOperator.configuration.helmCharts.jobImage.repository}}:{{.Values.agentOperator.configuration.helmCharts.jobImage.tag | default (include "image-tag" .) }}
{{- end -}}
