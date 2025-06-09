{{- /*gotype: github.com/kloudlite/operator/operators/app-n-lambda/internal/templates.HPATemplateVars*/ -}}
{{ with . }}
{{- if not .HPA }}
{{ fail ".HPA must be set" }}
{{- end }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata: {{.Metadata | toYAML | nindent 4}}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{.Metadata.Name}}
  minReplicas: {{ .HPA.MinReplicas }}
  maxReplicas: {{ .HPA.MaxReplicas }}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{.HPA.ThresholdCpu}}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{.HPA.ThresholdMemory}}
{{ end }}
