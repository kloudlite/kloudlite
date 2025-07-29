apiVersion: apps/v1
kind: Deployment
metadata: {{.Metadata | toJson }}
spec:
  selector: 
    matchLabels: {{.Selector | toJson }}
  {{- if .Paused }}
  paused: {{.Paused}}
  {{- end }}

  replicas: {{.Replicas | int}}
  template:
    metadata:
      labels: 
        {{- range $k, $v := .Selector }}
        {{$k}}: {{$v}}
        {{- end }}
        {{- range $k, $v := .PodLabels }}
        {{$k}}: {{$v}}
        {{- end }}
      annotations: {{.PodAnnotations | toJson}}
    spec: {{ .PodSpec | toJson }}

