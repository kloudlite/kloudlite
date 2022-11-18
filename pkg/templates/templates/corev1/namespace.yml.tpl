{{- $name := get . "name"  -}}
{{- $ownerRefs := get . "owner-refs"  -}}
{{- $labels := get . "labels"  -}}
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{$name}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
  labels: {{$labels | toYAML | nindent 4}}
