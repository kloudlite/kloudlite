{{- define "helmchart-operator-env" -}}
- name: HELM_JOB_RUNNER_IMAGE
  value: {{.Values.operators.platformOperator.configuration.helmCharts.jobImage.repository}}:{{.Values.operators.platformOperator.configuration.helmCharts.jobImage.tag | default (include "image-tag" .) }}
{{- end -}}

