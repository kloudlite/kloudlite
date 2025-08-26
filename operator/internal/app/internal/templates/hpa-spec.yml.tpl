scaleTargetRef:
  apiVersion: apps/v1
  kind: Deployment
  name: {{.DeploymentName}}
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
