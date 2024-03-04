{{- $apiVersion := get . "api-version" }}
{{- $kind := get . "kind" }}

{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}
{{- $labels := get . "labels" }}
{{- $ownerRefs := get . "owner-refs" }}

{{- $nodeSelector := get . "node-selector" }}
{{- $tolerations := get . "tolerations" }}

{{- $serviceTemplateSpec := get . "service-template-spec" }}

apiVersion: {{$apiVersion}}
kind: {{$kind}}
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{ $labels | toYAML | nindent 4 }}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
spec: 
  nodeSelector: {{$nodeSelector |toYAML | nindent 2}}
  tolerations: {{$tolerations |toYAML | nindent 2}}
  {{$serviceTemplateSpec | toYAML | nindent 2 }}
