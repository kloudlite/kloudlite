{{- define "msvc-n-mres-operator-env" -}}
- name: MSVC_CREDS_SVC_HTTP_PORT
  value: "{{.Values.operators.agentOperator.configuration.msvc.credsSvc.httpPort}}"

- name: MSVC_CREDS_SVC_REQUEST_PATH
  value: "{{.Values.operators.agentOperator.configuration.msvc.credsSvc.requestPath}}"
{{- end -}}
