{{- $klCloudflareWildcardDomains := get . "kl-cloudflare-wildcard-domains" | default list -}}
{{- $klCloudflareEmail := get . "kl-cloudflare-email" -}}
{{- $klCloudflareSecretName := get . "kl-cloudflare-secret-name" -}}

{{- $klAcmeEmail := get . "kl-acme-email" -}}

{{- $issuerName := get . "issuer-name" }}
{{- $ingressClass := get . "ingress-class" -}}

{{- $tolerations := get . "tolerations" | default list -}}
{{- $nodeSelector := get . "node-selector"  |default dict -}}
{{- $ownerRefs := get . "owner-refs"  -}}

apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: {{$issuerName}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4 }}
spec:
  acme:
    email: {{$klAcmeEmail}}
    privateKeySecretRef:
      name: {{$issuerName}}
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      - dns01:
          cloudflare:
            email: {{$klCloudflareEmail}}
            apiTokenSecretRef:
              name: {{$klCloudflareSecretName}}
              key: api-token
        selector:
          dnsNames: {{ $klCloudflareWildcardDomains | toYAML | nindent 12 }}
{{/*            - "*.$DOMAIN_1"*/}}
{{/*            - "*.$DOMAIN_2"*/}}
{{/*            - "crewscale.kl-client.kloudlite.io"*/}}
{{/*            - "*.crewscale.kl-client.kloudlite.io"*/}}
      - http01:
          ingress:
            class: {{$ingressClass}}
            podTemplate:
              spec:
                {{if $nodeSelector}}
                nodeSelector: {{$nodeSelector | toYAML | nindent 18 }}
                {{end}}
                {{if $tolerations}}
                tolerations: {{$tolerations | toYAML | nindent 18 }}
                {{end}}
