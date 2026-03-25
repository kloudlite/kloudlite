{{/*
Resolve image tag: use per-app tag if set, otherwise global.imageTag
*/}}
{{- define "kloudlite.imageTag" -}}
{{- . | default $.Values.global.imageTag -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "kloudlite.labels" -}}
app.kubernetes.io/managed-by: Helm
app.kubernetes.io/part-of: kloudlite
{{- end -}}
