{{- $ownerRefs := get . "owner-refs" }}
{{- $obj := get . "object"}}
{{- with $obj }}
{{- /* gotype: github.com/kloudlite/operator/apis/crds/v1.ManagedResource */ -}}
apiVersion: {{.Spec.MsvcRef.APIVersion}}
kind: {{.Spec.MresKind.Kind}}
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
  labels: {{.Labels | default dict | toYAML | nindent 4}}
spec:
  msvcRef: {{.Spec.MsvcRef |toYAML |nindent 4}}
  {{- if .Spec.Inputs }}
  {{.Spec.Inputs | toYAML | nindent 2 }}
  {{- end}}
{{- end}}
