{{- $name := get . "name"  -}}
{{- $kubeconfigName := get . "kubeconfig-name"  -}}
{{- $kubeconfigNamespace := get . "kubeconfig-namespace"  -}}

apiVersion: extensions.kloudlite.io/v1
kind: Cluster
metadata:
  name: {{$name}}
spec:
  kubeConfig:
    name: {{$kubeconfigName}}
    namespace: {{$kubeconfigNamespace}}
