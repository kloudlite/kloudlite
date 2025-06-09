{{- $apiVersion := get . "api-version" }}
{{- $kind := get . "kind" }}

{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}

{{- $ownerRefs := get . "owner-refs" }}

{{- $labels := get . "labels" }}
{{- $annotations := get . "annotations" }}

{{- $serviceTemplateSpec := get . "service-template-spec" }}

{{- /* {{- $output := get . "output" }} */}}

{{- $export := get . "export" }}

apiVersion: {{$apiVersion}}
kind: {{$kind}}
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{ $labels | toYAML | nindent 4 }}
  annotations: {{ $annotations | toYAML | nindent 4 }}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
spec: 
  {{$serviceTemplateSpec | toYAML | nindent 2 }}

{{- if $export }}
export:
  {{- if $export.ViaSecret }}
  viaSecret: {{ $export.ViaSecret }}
  {{- end }}

  {{- if $export.Template }}
  template: {{ $export.Template }}
  {{- end }}
{{- end }}

{{- /* {{- if $output }} */}}
{{- /* output: {{$output | toYAML | nindent 2}} */}}
{{- /* {{- end }} */}}
