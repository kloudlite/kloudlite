{{- if .Values.cloudflareWildcardCert.enabled }}
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{.Values.cloudflareWildcardCert.name}}
  {{/* namespace: {{.Release.Namespace}} */}}
  namespace: {{.Values.operatorsNamespace}}
spec:
  dnsNames:
  {{range $v := .Values.cloudflareWildcardCert.domains}}
    - {{$v | squote}}
  {{end}}
  secretName: {{.Values.cloudflareWildcardCert.secretName}}
  issuerRef:
    name: {{.Values.clusterIssuer.name}}
    kind: ClusterIssuer
{{- end}}
