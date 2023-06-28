{{- define "serviceAccountName" -}}
{{- printf "%s-%s" .Release.Name .Values.svcAccountName -}}
{{- end}}
