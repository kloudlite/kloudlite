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

{{- define "gcp-credentials-secret-name" -}}
{{$.Release.Name}}-{{$.Values.gcp.gcloudServiceAccountCreds.nameSuffix}}
{{- end -}}

{{- define "gcp-csi-namespace" -}}
gce-pd-csi-driver
{{- end -}}

{{- define "image-tag" -}}
{{ .Values.kloudliteRelease | default .Chart.AppVersion }}
{{- end -}}

