{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}
{{- $externalName := get . "external-name"}}
{{- $ownerRefs := get . "owner-refs"}}
{{- $labels := get . "labels"}}

apiVersion: v1
kind: Service
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  {{- if $labels }}
  labels: {{$labels | toYAML | nindent 4}}
  {{- end }}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
spec:
  type: ExternalName
  externalName: {{$externalName}}
