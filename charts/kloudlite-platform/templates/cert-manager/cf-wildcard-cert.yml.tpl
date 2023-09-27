{{- if .Values.cloudflareWildCardCert.create }}
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.cloudflareWildCardCert.name}}-cf-api-token
  namespace: {{.Release.Namespace}}
stringData:
  api-token: {{.Values.cloudflareWildCardCert.cloudflareCreds.secretToken}}

---

apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{.Values.cloudflareWildCardCert.name}}
  namespace: {{.Release.Namespace}}
spec:
  dnsNames:
  {{range $v := .Values.cloudflareWildCardCert.domains}}
    - {{$v | squote}}
  {{end}}
  secretName: {{.Values.cloudflareWildCardCert.secretName}}
  issuerRef:
    name: {{.Values.clusterIssuer.name}}
    kind: ClusterIssuer
{{- end}}
