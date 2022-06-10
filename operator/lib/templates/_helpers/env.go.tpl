{{- define "TemplateEnv" }}
{{- range $_, $v := . }}
{{- with $v }}
  - name: {{.Key}}
  {{- if .Value }}
    value: {{.Value | squote }}
  {{- else }}
    valueFrom:
    {{- if eq .Type "config" }}
      configMapKeyRef:
      name: {{.RefName}}
      key: {{.RefKey}}
    {{- end }}
    {{- if eq .Type "secret" }}
      secretKeyRef:
      name: {{.RefName}}
      key: {{.RefKey}}
    {{- end }}
  {{- end }}
{{- end}}
{{- end }}
{{- end }}
