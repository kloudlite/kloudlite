{{- define "resource-watcher-env" -}}
- name: GRPC_ADDR
  value: {{.Values.messageOfficeGRPCAddr}}

- name: ACCESS_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{.Values.clusterIdentitySecretName}}
      key: ACCESS_TOKEN
      optional: true
{{- end -}}
