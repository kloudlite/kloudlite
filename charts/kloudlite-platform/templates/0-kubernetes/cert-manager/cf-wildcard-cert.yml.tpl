{{- if .Values.cloudflareWildCardCert.enabled }}
apiVersion: v1
kind: Secret
metadata:
  name: kloudlite-cf-api-token
  namespace: {{.Release.Namespace}}
stringData:
  api-token: {{.Values.cloudflareWildCardCert.cloudflareCreds.secretToken}}

---

apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{.Values.cloudflareWildCardCert.certificateName}}
  namespace: {{.Release.Namespace}}
spec:
  dnsNames:
  {{range $v := .Values.cloudflareWildCardCert.domains}}
    - {{$v | squote}}
  {{end}}
  secretName: kl-cert-wildcard-tls
  issuerRef:
    name: {{.Values.certManager.certIssuer.name}}
    kind: ClusterIssuer
{{- end}}
