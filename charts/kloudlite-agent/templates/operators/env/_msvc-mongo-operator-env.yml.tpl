{{- define "msvc-mongo-operator-env" -}}
{{- /* - name: CLUSTER_INTERNAL_DNS */}}
{{- /*   value: "{{.Values.clusterInternalDNS}}" */}}
{{- /**/}}

- name: GLOBAL_VPN_DNS
  value: "{{.Values.clusterName}}.local"

- name: MSVC_CREDS_SVC_NAME
  value: "{{.Values.operators.agentOperator.configuration.msvc.credsSvc.name}}"

- name: MSVC_CREDS_SVC_NAMESPACE
  value: "{{.Release.Namespace}}"
{{- end -}}
