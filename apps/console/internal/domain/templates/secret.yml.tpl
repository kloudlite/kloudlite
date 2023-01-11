{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $kubeConfig := get . "kubeconfig"  -}}

apiVersion: v1
kind: Secret
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels:
    kloudlite.io/cluster-config: "true"
stringData:
  kubeConfig: |+
    {{$kubeConfig | nindent 4 }}
