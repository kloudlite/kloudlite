{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}
{{- $ownerRefs := get . "owner-refs" }}
{{- $labels := get . "labels" }}
{{- $data := get . "data" }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
  {{- if $labels }}
  labels: {{$labels | toYAML | nindent 4}}
  {{- end }}
{{- if $data }}
data: {{$data | toYAML | nindent 2}}
{{- end }}
