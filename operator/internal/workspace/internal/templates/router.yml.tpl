apiVersion: kloudlite.io/v1
kind: Router
metadata: {{.Metadata | toJson}}
spec:
  https:
    enabled: true
    forceRedirect: true
  routes:
    {{- if .EnableCodeServer }}
    - host: code-{{.Metadata.Name}}-{{.WorkMachineName}}.{{.KloudliteDomain}}
      service: {{.ServiceName}}
      path: {{.ServicePath}}
      port: {{.CodeServerPort}}
    {{- end }}

    {{- if .EnableJupyterNotebook }}
    - host: notebook-{{.Metadata.Name}}-{{.WorkMachineName}}.{{.KloudliteDomain}}
      service: {{.ServiceName}}
      path: {{.ServicePath}}
      port: {{.NotebookPort}}
    {{- end }}

    {{- if .EnableTTYD }}
    - host: ttyd-{{.Metadata.Name}}-{{.WorkMachineName}}.{{.KloudliteDomain}}
      service: {{.ServiceName}}
      path: {{.ServicePath}}
      port: {{.TTYDPort}}
    {{- end }}
