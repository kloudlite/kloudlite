{{- if .Values.certManager.enabled }}
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: {{.Values.certManager.certIssuer.name}}
spec:
  acme:
    email: {{.Values.certManager.certIssuer.acmeEmail}}
    privateKeySecretRef:
      name: {{.Values.certManager.certIssuer.name}}
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      {{- if .Values.cloudflareWildCardCert.enabled}}
      - dns01:
          cloudflare:
            email: {{.Values.cloudflareWildCardCert.cloudflareCreds.email}}
            apiTokenSecretRef:
              name: kloudlite-cf-api-token
              key: api-token
        selector:
          dnsNames:
            {{- range $v := .Values.cloudflareWildCardCert.domains}}
            - {{$v | squote}}
            {{- end }}
      {{- end}}
      {{- $ingClass := .Values.global.ingressClassName }}
      {{- if $ingClass }}
      - http01:
          ingress:
            class: "{{$ingClass}}"
      {{- end}}
{{- end }}