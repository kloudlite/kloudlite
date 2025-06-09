{{- $apiVersion := get . "api-version" }}
{{- $kind := get . "kind" }}

{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}
{{- $ownerRefs := get . "owner-refs" }}
{{- $labels := get . "labels" }}

{{- $msvcRef := get . "msvc-ref" }}
{{- $resourceTemplateSpec := get . "resource-template-spec" }}

{{- $export := get . "export" }}

apiVersion: {{$apiVersion}}
kind: {{$kind}}
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
  labels: {{ $labels | toYAML | nindent 4}}
spec:
  managedServiceRef: {{$msvcRef | toYAML | nindent 4 }}
  {{- if $resourceTemplateSpec }}
  {{ $resourceTemplateSpec | toYAML | nindent 2 }}
  {{- end}}

{{- /* {{- if $output }} */}}
{{- /* output: {{$output | toYAML | nindent 2 }} */}}
{{- /* {{- end }} */}}

{{- if $export }}
export:
  {{- if $export.ViaSecret }}
  viaSecret: {{ $export.ViaSecret }}
  {{- end }}

  {{- if $export.Template }}
  template: {{ $export.Template }}
  {{- end }}
{{- end }}
