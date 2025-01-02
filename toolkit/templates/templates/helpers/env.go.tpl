{{- define "TemplateEnv" }}
{{- range $_, $v := . }}
{{- with $v }}
- name: {{.Key | squote}}
{{- if .Value }}
  value: {{.Value | squote }}
{{- else }}
  valueFrom:
  {{- if eq .Type "config" }}
    configMapKeyRef:
      name: {{.RefName}}
      key: {{.RefKey}}
      optional: {{.Optional | default false}}
  {{- end }}
  {{- if eq .Type "secret" }}
    secretKeyRef:
      name: {{.RefName}}
      key: {{.RefKey}}
      optional: {{.Optional | default false}}
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
