{{- $apiVersion := get . "api-version" }}
{{- $kind := get . "kind" }}

{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}
{{- $labels := get . "labels" }}
{{- $ownerRefs := get . "owner-refs" }}

{{- $serviceTemplateSpec := get . "service-template-spec" }}

apiVersion: {{$apiVersion}}
kind: {{$kind}}
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{ $labels | toYAML | nindent 4 }}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
spec: {{$serviceTemplateSpec | toYAML | nindent 2 }}
