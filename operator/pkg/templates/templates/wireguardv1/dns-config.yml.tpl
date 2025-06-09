{{- $rewriteRules := get . "rewrite-rules"}}
{{- $namespace := get . "namespace"}}
{{- $name := get . "name"}}
{{- $labels := get . "labels"}}
{{- $ownerRefs := get . "ownerRefs"}}

apiVersion: v1
data:
  Corefile: |+
{{ $rewriteRules | indent 4 }}
kind: ConfigMap
metadata:
  name: "wg-dns-{{ $name }}"
  namespace: {{ $namespace }}
  labels: {{ $labels | toJson }}
  ownerReferences: {{ $ownerRefs| toJson}}
