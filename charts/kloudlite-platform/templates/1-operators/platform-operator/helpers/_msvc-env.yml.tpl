{{- define "msvc-operator-env" -}}
- name: MSVC_CREDS_SVC_NAME
  value: "test"

- name: MSVC_CREDS_SVC_NAMESPACE
  value: "adafsd"

- name: "MSVC_CREDS_SVC_HTTP_PORT"
  value: "8001"

- name: MSVC_CREDS_SVC_REQUEST_PATH
  value: "/get-msvc-creds"
{{- end -}}

