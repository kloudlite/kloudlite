{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $labels := get . "labels" | default dict -}}
{{- $ownerRefs := get . "owner-refs" | default list  -}}
{{- $data := get . "data" | default dict -}}
{{- $stringData := get . "string-data" | default dict -}}
{{- $secretType := get . "secret-type" | default "Opaque" -}}
{{- $immutable := get . "immutable" | default false -}}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  {{- if $ownerRefs }}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4 }}
  {{- end }}
  labels: {{$labels | toYAML | nindent 4}}
type: {{$secretType}}
stringData: {{ $stringData | toYAML | nindent 2 }}
data: {{ $data | toYAML | nindent 2 }}
immutable: {{ $immutable }}
