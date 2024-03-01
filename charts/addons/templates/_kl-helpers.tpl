{{- define "node-selector-masters" -}}
node-role.kubernetes.io/master: "true"
{{- end -}}

{{- define "node-tolerations-masters" -}}
- key: node-role.kubernetes.io/master
  operator: Exists
{{- end -}}

{{- define "node-selector-agent" -}}
kloudlite.io/node.has-role: agent
{{- end -}}
