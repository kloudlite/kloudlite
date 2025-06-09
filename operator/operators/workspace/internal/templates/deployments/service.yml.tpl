---
{{- with . }}
apiVersion: v1
kind: Service
metadata: {{.Metadata | toJson }}
spec:
  selector:
    app: {{.Metadata.Name | squote}}
  ports:
    - name: "ssh"
      protocol: "TCP"
      port: {{.PortConfig.SSHPort}}
      targetPort: {{.PortConfig.SSHPort}}

{{ if .EnableTTYD }}
    - name: "ttyd-server"
      protocol: "TCP"
      port: {{.PortConfig.TTYDPort}}
      targetPort: {{.PortConfig.TTYDPort}}
{{ end }}
    
{{ if .EnableJupyterNotebook }}
    - name: "jupyter-server"
      protocol: "TCP"
      port: {{.PortConfig.NotebookPort}}
      targetPort: {{.PortConfig.NotebookPort}}
{{ end }}

{{ if .EnableCodeServer }}
    - name: "code-server"
      protocol: "TCP"
      port: {{.PortConfig.CodeServerPort}}
      targetPort: {{.PortConfig.CodeServerPort}}
{{ end }}


---
apiVersion: v1
kind: Service
metadata:
  name: {{.Metadata.Name}}-headless
  namespace: {{.Metadata.Namespace}}
  labels: {{.Metadata.Labels | toJson }}
  annotations: {{.Metadata.Annotations | toJson }}
  ownerReferences: {{.Metadata.OwnerReferences | toJson }}
spec:
  clusterIP: None
  selector:
    app: {{.Metadata.Name | squote}}

{{- end }}