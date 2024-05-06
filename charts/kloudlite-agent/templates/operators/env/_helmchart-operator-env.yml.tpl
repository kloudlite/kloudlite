{{- define "helmchart-operator-env" -}}
- name: HELM_JOB_RUNNER_IMAGE
  value: {{.Values.operators.agentOperator.configuration.helmCharts.jobImage.repository}}:{{.Values.operators.agentOperator.configuration.helmCharts.jobImage.tag | default (include "image-tag" .) }}
{{- end -}}
