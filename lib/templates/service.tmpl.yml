{{- $ownerRefs := get . "owner-refs"}}
{{- $obj := get . "object" }}

{{- with $obj}}
apiVersion: v1
kind: Service
metadata:
  namespace: {{.Namespace}}
  name: {{.Name}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
spec:
  selector:
    app: {{.Name}}
  ports:
    {{- range $svc := .Spec.Services }}
    {{- with $svc }}
    - protocol: {{.Type | upper | default "TCP"}}
      port: {{.Port}}
      name: {{.Name}}
      targetPort: {{.TargetPort}}
    {{- end }}
    {{- end }}
{{- end}}
