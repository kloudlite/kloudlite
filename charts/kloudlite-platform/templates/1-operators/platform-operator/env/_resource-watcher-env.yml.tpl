{{- define "resource-watcher-env" -}}
- name: ACCOUNT_NAME
  value: {{.Values.global.accountName}}

- name: CLUSTER_NAME
  value: {{.Values.global.clusterName}}

- name: PLATFORM_ACCESS_TOKEN
  value: {{.Values.apps.messageOfficeApi.configuration.platformAccessToken}}

- name: GRPC_ADDR
{{/*TODO check with anshuman*/}}
  value: {{.Values.apps.messageOfficeApi.name}}:{{.Values.apps.messageOfficeApi.configuration.externalGrpcPort}}

{{- end -}}
