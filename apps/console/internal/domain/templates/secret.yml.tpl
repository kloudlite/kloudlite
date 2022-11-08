{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $kubeConfig := get . "kubeconfig"  -}}

apiVersion: v1
kind: Secret
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
stringData:
  kubeConfig: |+
    {{$kubeConfig | nindent 4 }}
