{{- $name := get . "name" -}}
{{- $prom := get . "prometheus-values" -}}
{{- $storageClass := get . "storage-class" -}}

{{- if not (or $prom $storageClass) -}}
{{- fail "prometheus-values, and storage-class must be defined " }}
{{- end -}}

fullnameOverride: {{$name}}
nameOverride: {{$name}}

{{- with $prom }}
{{/*gotype: operators.kloudlite.io/apis/cluster-setup/v1.PrometheusValues*/}}
operator:
  resources:
    requests:
      cpu: 100m
      memory: 200Mi
    limits:
      cpu: 100m
      memory: 200Mi

prometheus:
  enabled: true
  persistence:
    enabled: true
    size: {{.Resources.Storage.Size}}
    storageClass: {{$storageClass}}

  resources:
    requests:
      cpu: {{.Resources.Cpu.Min}}
      memory: {{.Resources.Memory}}
    limits:
      cpu: {{.Resources.Cpu.Max}}
      memory: {{.Resources.Memory}}

  livenessProbe:
    enabled: true
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 7

  readinessProbe:
    enabled: true
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 7
{{- end }}
