{{- define "msvc-mongo-operator-env" -}}
{{- /* - name: CLUSTER_INTERNAL_DNS */}}
{{- /*   value: "{{.Values.clusterInternalDNS}}" */}}
{{- /**/}}

- name: GLOBAL_VPN_DNS
  value: "{{.Values.clusterName}}.local"
{{- end -}}
