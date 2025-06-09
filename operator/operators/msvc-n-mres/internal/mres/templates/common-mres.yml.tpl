{{ with . }}
apiVersion: {{.PluginTemplate.APIVersion}}
kind: {{.PluginTemplate.Kind}}
metadata: {{.Metadata | toJson }}
spec:
  managedServiceRef: {{ .ManagedServiceRef | toYAML | nindent 4}}
  {{.PluginTemplate.Spec | toYAML | nindent 2 }}
{{- if .PluginTemplate.Export }}
export:
  {{- if .PluginTemplate.Export.ViaSecret }}
  viaSecret: {{ .PluginTemplate.Export.ViaSecret }}
  {{- end }}

  {{- if .PluginTemplate.Export.Template }}
  template: |+
    {{- .PluginTemplate.Export.Template | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}
