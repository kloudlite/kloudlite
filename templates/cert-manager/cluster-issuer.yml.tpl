{{ if .Values.cloudflareWildcardCert.enabled }}
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.cloudflareWildcardCert.name}}-cf-api-token
  namespace: {{.Values.operatorsNamespace}}
stringData:
  api-token: {{.Values.cloudflareWildcardCert.cloudflareCreds.secretToken}}
{{ end }}

---

apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: {{.Values.clusterIssuer.name}}
spec:
  acme:
    email: {{.Values.clusterIssuer.acmeEmail}}
    privateKeySecretRef:
      name: {{.Values.clusterIssuer.name}}
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      {{- if .Values.cloudflareWildcardCert.enabled}}
      - dns01:
          cloudflare:
            email: {{.Values.cloudflareWildcardCert.cloudflareCreds.email}}
            apiTokenSecretRef:
              name: {{.Values.cloudflareWildcardCert.name}}-cf-api-token
              key: api-token
        selector:
          dnsNames:
            {{- range $v := .Values.cloudflareWildcardCert.domains}}
            - {{$v | squote}}
            {{- end }}
      {{- end}}
      {{- if .Values.ingressClassName }}
      - http01:
          ingress:
            class: "{{.Values.ingressClassName}}"
      {{- end}}
