---
{{- with . }}
apiVersion: v1
kind: Service
metadata: {{.Metadata | toJson }}
spec:
  selector: {{.Selector | toJson}}
  ports:
    - name: "ssh"
      protocol: "TCP"
      port: {{.SSHPort}}
      targetPort: {{.SSHPort}}

{{ if .EnableTTYD }}
    - name: "ttyd-server"
      protocol: "TCP"
      port: {{.TTYDPort}}
      targetPort: {{.TTYDPort}}
{{ end }}
    
{{ if .EnableJupyterNotebook }}
    - name: "jupyter-server"
      protocol: "TCP"
      port: {{.NotebookPort}}
      targetPort: {{.NotebookPort}}
{{ end }}

{{ if .EnableVSCodeServer }}
    - name: "vscode-server"
      protocol: "TCP"
      port: {{.VSCodeServerPort}}
      targetPort: {{.VSCodeServerPort}}
{{ end }}

---

apiVersion: v1
kind: Service
metadata: {{.HeadlessServiceMetadata | toJson }}
spec:
  clusterIP: None
  selector: {{.Selector | toJson}}

{{- end }}
