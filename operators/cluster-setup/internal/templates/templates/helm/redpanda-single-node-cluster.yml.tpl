{{- $namespace := get . "namespace" -}}
{{- $labels := get . "labels" | default dict -}}

{{- $clusterName := get . "cluster-name" -}}
{{- $version := get . "version" | default "v22.1.6" -}}
{{- $adminUsername := get . "admin-username" | default "admin" -}}
{{- $baseDomain := get . "base-domain" -}}
{{/*{{- $kafkaSubDomain := get . "kafka-sub-domain" | default "kafka.kloudlite.io" -}}*/}}
  
{{- $storageSize := get . "storage-size" -}}
{{- $storageClass := get . "storage-class" -}}

{{- $nodeSelector := get . "node-selector" | default dict -}}
{{- $tolerations := get . "tolerations" | default list -}}

apiVersion: redpanda.vectorized.io/v1alpha1
kind: Cluster
metadata:
  name: {{$clusterName}}
  namespace: {{$namespace}}
  labels: {{$labels | toYAML  | nindent 4}}
spec:
  image: "vectorized/redpanda"
  version: {{$version}}
  replicas: 1
  enableSasl: true
  superUsers:
    - username: {{$adminUsername}}
  resources:
    requests:
      cpu: 200m
      memory: 300Mi
    limits:
      cpu: 300m
      memory: 300Mi
  nodeSelector: {{$nodeSelector | toYAML | nindent 4}}
  tolerations: {{$tolerations | toYAML | nindent 4}}
  configuration:
    rpcServer:
      port: 33145
    kafkaApi:
      - port: 9092
      - external:
          enabled: true
          subdomain: "kafka.{{$baseDomain}}"
    pandaproxyApi:
      - port: 8082
      - external:
          enabled: true
    schemaRegistry:
      port: 8081
      external:
        enabled: true
    adminApi:
      - port: 9644
    developerMode: true
  storage:
    capacity: {{$storageSize}}
{{/*    XFS*/}}
    storageClassName: {{$storageClass}}
