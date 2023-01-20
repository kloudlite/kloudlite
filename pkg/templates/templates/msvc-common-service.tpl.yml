{{- $ownerRefs := get . "owner-refs" -}}
{{- $obj := get . "obj" -}}
{{- with $obj }}
{{- /*gotype:  github.com/kloudlite/operator/apis/crds/v1.ManagedService */ -}}
apiVersion: {{.Spec.MsvcKind.APIVersion}}
kind: {{.Spec.MsvcKind.Kind}}
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels: {{ .Labels | toYAML | nindent 4 }}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
spec:
  {{- if .Spec.NodeSelector }}
  nodeSelector: {{ .Spec.NodeSelector | toYAML | nindent 4 }}
  {{- end }}
  region: {{.Spec.Region}}
  {{ .Spec.Inputs | toYAML | nindent 2 }}
{{- end}}
