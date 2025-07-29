{{- with . }}
apiVersion: {{.PluginTemplate.APIVersion}}
kind: {{.PluginTemplate.Kind}}
metadata: {{.Metadata | toYAML | nindent 2 }}
spec: {{.PluginTemplate.Spec | toYAML | nindent 2 }}
export:
  {{- if .PluginTemplate.Export.ViaSecret }}
  viaSecret: {{ .PluginTemplate.Export.ViaSecret }}
  {{- end }}

  {{- if .PluginTemplate.Export.Template }}
  template: |+ 
    {{- .PluginTemplate.Export.Template | nindent 4 }}
  {{- end }}
{{- end }}
