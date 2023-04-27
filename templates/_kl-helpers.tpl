{{- define "node-selector-and-tolerations" -}}
{{- if .Values.nodeSelector -}}
nodeSelector: {{ .Values.nodeSelector | toYaml | nindent 2 }}
{{- end }}

{{- if .Values.tolerations -}}
tolerations: {{ .Values.tolerations | toYaml | nindent 2 }}
{{- end -}}
{{- end -}}
