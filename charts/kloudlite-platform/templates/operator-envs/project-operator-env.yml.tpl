{{- define "project-operator-env" -}}
- name: SVC_ACCOUNT_NAME
  value: "kloudlite-svc-account"

- name: CLUSTER_INTERNAL_DNS
  value: "{{.Values.clusterInternalDNS}}"
{{- end -}}
