apiVersion: apps/v1
kind: Deployment
metadata: {{.Metadata | toJson }}
spec:
  selector: 
    matchLabels:
      app: {{.Metadata.Name}}
  {{- if .Paused }}
  paused: {{.Paused}}
  {{- end }}

  replicas: {{.Replicas | int}}
  template:
    metadata:
      labels: 
        app: {{.Metadata.Name}}
        {{- range $k, $v := .PodLabels }}
        {{$k}}: {{$v}}
        {{- end }}
      annotations: {{.PodAnnotations | toJson}}
    spec: {{ .PodSpec | toJson }}

