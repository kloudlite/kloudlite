{{- if .Values.clusterIssuer.create }}
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
      {{- if .Values.cloudflareWildCardCert.create}}
      - dns01:
          cloudflare:
            email: {{.Values.cloudflareWildCardCert.cloudflareCreds.email}}
            apiTokenSecretRef:
              name: {{.Values.cloudflareWildCardCert.name}}-cf-api-token
              key: api-token
        selector:
          dnsNames:
            {{- range $v := .Values.cloudflareWildCardCert.domains}}
            - {{$v | squote}}
            {{- end }}
      {{- end}}
      {{- $ingClass := (index .Values.helmCharts "ingress-nginx").configuration.ingressClassName }} 
      {{- if $ingClass }}
      - http01:
          ingress:
            class: "{{$ingClass}}"
      {{- end}}
{{- end }}

apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: cluster-issuer
spec:
  acme:
    email: support@kloudlite.io
    privateKeySecretRef:
      name: cluster-issuer-tls
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      - http01:
          ingress:
            class: "ingress-nginx"
