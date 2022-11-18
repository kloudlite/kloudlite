{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $labels := get . "labels" -}}
{{- $ownerRefs := get . "owner-refs"  -}}
{{- $data := get . "data"  -}}
{{- $immutable := get . "immutable"  -}}

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
  {{- if $labels }}
  labels: {{$labels | toYAML | nindent 4}}
  {{- end}}
data: {{$data | toYAML | nindent 2}}
immutable: {{$immutable}}
