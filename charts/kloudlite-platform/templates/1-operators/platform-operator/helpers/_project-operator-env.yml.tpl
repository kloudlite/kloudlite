{{- define "project-operator-env" -}}
- name: CLUSTER_INTERNAL_DNS
  value: "{{.Values.global.clusterInternalDNS}}"
{{- end -}}
