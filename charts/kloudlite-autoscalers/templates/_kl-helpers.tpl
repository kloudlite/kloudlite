{{- define "service-account-name" -}}
{{.Release.Name}}-{{.Values.serviceAccount.nameSuffix}}
{{- end -}}

{{- define "image-tag" -}}
{{ .Values.kloudliteRelease | default .Chart.AppVersion }}
{{- end -}}

{{- define "image-pull-policy" -}}
{{- if .Values.imagePullPolicy -}}
{{- .Values.imagePullPolicy}}
{{- else -}}
{{- if hasSuffix "-nightly" (include "image-tag" .) -}}
{{- "Always" }}
{{- else -}}
{{- "IfNotPresent" }}
{{- end -}}
{{- end -}}
{{- end -}}
