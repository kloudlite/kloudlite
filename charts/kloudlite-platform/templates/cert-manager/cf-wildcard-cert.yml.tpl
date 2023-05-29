{{- if .Values.clusterIssuer.cloudflareWildCardCert.create }}
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{.Values.clusterIssuer.cloudflareWildCardCert.name}}
  namespace: {{.Values.operatorsNamespace}}
spec:
  dnsNames:
  {{range $v := .Values.clusterIssuer.cloudflareWildCardCert.domains}}
    - {{$v | squote}}
  {{end}}
  secretName: {{.Values.clusterIssuer.cloudflareWildCardCert.secretName}}
  issuerRef:
    name: {{.Values.clusterIssuer.name}}
    kind: ClusterIssuer
{{- end}}
