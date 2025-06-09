{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $accRef := get . "acc-ref"  -}}
---
apiVersion: crds.kloudlite.io/v1
kind: AccountRouter
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
spec:
  accountRef: {{$accRef}}
  serviceType: ClusterIP
  https:
    enabled: true
    forceRedirect: true
