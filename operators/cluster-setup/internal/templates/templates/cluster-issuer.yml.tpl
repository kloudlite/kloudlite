{{- $clusterIssuer := get . "cluster-issuer" -}}

{{- with $clusterIssuer }}
{{/*gotype: operators.kloudlite.io/apis/cluster-setup/v1.clusterIssuer*/}}
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: {{.Name}}
spec:
  acme:
    email: {{.AcmeEmail}}
    preferredChain: ""
    privateKeySecretRef:
      name: {{.Name}}
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      - dns01:
          cloudflare:
            apiTokenSecretRef:
            {{/*              key: api-token*/}}
            {{/*              name: cloudflare-zone-secret*/}}
              key: {{.Cloudflare.SecretKeyRef.Key}}
              name: {{.Cloudflare.SecretKeyRef.Name}}
            email: {{.Cloudflare.Email}}
        selector:
          dnsNames: {{.Cloudflare.DnsNames | toYAML | nindent 7 }}
{{/*            - '*.kloudlite.io'*/}}
{{/*            - '*.khost.dev'*/}}
{{/*            - crewscale.kl-client.kloudlite.io*/}}
{{/*            - '*.crewscale.kl-client.kloudlite.io'*/}}
      - http01:
          ingress:
            class: {{.IngressClass}}
