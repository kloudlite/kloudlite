{{- if .Values.clusterIssuer.create }}

{{ if .Values.clusterIssuer.cloudflareWildCardCert.create }}
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.clusterIssuer.cloudflareWildCardCert.name}}-cf-api-token
  namespace: {{.Values.operatorsNamespace}}
stringData:
  api-token: {{.Values.clusterIssuer.cloudflareWildCardCert.cloudflareCreds.secretToken}}
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
      {{- if .Values.clusterIssuer.cloudflareWildCardCert.create}}
      - dns01:
          cloudflare:
            email: {{.Values.clusterIssuer.cloudflareWildCardCert.cloudflareCreds.email}}
            apiTokenSecretRef:
              name: {{.Values.clusterIssuer.cloudflareWildCardCert.name}}-cf-api-token
              key: api-token
        selector:
          dnsNames:
            {{- range $v := .Values.clusterIssuer.cloudflareWildCardCert.domains}}
            - {{$v | squote}}
            {{- end }}
      {{- end}}
      {{- if .Values.ingressClassName }}
      - http01:
          ingress:
            class: "{{.Values.ingressClassName}}"
      {{- end}}

{{- end }}
