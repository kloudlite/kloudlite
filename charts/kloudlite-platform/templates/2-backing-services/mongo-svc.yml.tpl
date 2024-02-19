{{- $name := "mongo-svc" }}

{{- if .Values.mongo.runAsCluster }}
apiVersion: mongodb.msvc.kloudlite.io/v1
kind: ClusterService
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: {{.Values.mongo.replicas}}
  tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 4}}
  nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 4}}
  resources:
    cpu:
      min: 300m
      max: 500m
    memory: 
      min: 500Mi
      max: 500Mi
    storage:
      size: {{.Values.mongo.configuration.volumeSize}}
      storageClass: sc-xfs
{{ else }}
apiVersion: mongodb.msvc.kloudlite.io/v1
kind: StandaloneService
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
spec:
  tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 4}}
  nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 4}}
  resources:
    cpu:
      min: 300m
      max: 500m
    memory: 
      min: 500Mi
      max: 500Mi
    storage:
      size: {{.Values.mongo.configuration.volumeSize}}
      storageClass: sc-xfs
{{- end }}
---
