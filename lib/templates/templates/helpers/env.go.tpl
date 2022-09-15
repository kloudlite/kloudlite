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

{{- define "KAnnotation" }}
{{- $items := . }}

{{- if eq (len $items) 3 }}
{{- if (index $items 0) }}
{{index $items 1}}: {{index $items 2 | squote}}
{{- end}}
{{- end}}

{{- if eq (len $items) 2 }}
{{index $items 1}}: {{index $items 2 | squote}}
{{- end}}

{{- end}}
