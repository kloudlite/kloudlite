{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $region := get . "region"  -}}
apiVersion: elasticsearch.msvc.kloudlite.io/v1
kind: Service
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
spec:
  region: {{$region}}
  replicaCount: 1
  resources:
    storage:
      size: 2Gi
    cpu:
      min: 610m
      max: 800m
    memory: 800Mi
