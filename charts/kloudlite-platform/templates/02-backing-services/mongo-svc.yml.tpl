{{- $name := "mongo-svc" }}

{{- if .Values.mongo.runAsCluster }}
apiVersion: mongodb.msvc.kloudlite.io/v1
kind: ClusterService
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: {{.Values.mongo.replicas}}
  nodeSelector: {{.Values.mongo.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.mongo.tolerations | toYaml | nindent 4}}
  resources:
    cpu:
      min: 300m
      max: 500m
    memory: 
      min: 500Mi
      max: 500Mi
    storage:
      size: {{.Values.mongo.volumeSize}}
      storageClass: sc-xfs
output:
  credentialsRef:
    name: msvc-{{$name}}-creds
{{ else }}
apiVersion: plugin-mongodb.kloudlite.github.com/v1
kind: StandaloneService
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
spec:
  nodeSelector: {{.Values.mongo.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.mongo.tolerations | toYaml | nindent 4}}
  resources:
    cpu:
      min: 300m
      max: 500m
    memory: 
      min: 500Mi
      max: 500Mi
    storage:
      size: {{.Values.mongo.volumeSize}}
      storageClass: sc-xfs
output:
  credentialsRef:
    name: msvc-{{$name}}-creds
{{- end }}
---
