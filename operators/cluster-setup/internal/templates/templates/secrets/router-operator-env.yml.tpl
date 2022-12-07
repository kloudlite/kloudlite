{{- $namespace := get . "namespace" -}}
{{- $clusterWildcardDomain := get . "cluster-wildcard-domain" -}}
{{- $acmeEmail := get . "acme-email" | default "support@kloudlite.io" -}}
{{- $cloudflareSecretName := get . "cloudflare-secret-name" -}}
{{- $cloudflareEmail := get . "cloudflare-email" -}}

apiVersion: v1
kind: Secret
metadata:
  name: "router-operator-env"
  namespace: {{$namespace}}
stringData:
  CLOUDFLARE_WILDCARD_DOMAINS: '{{$clusterWildcardDomain}}'
  CLOUDFLARE_EMAIL: '{{$cloudflareEmail}}'
  CLOUDFLARE_SECRET_NAME: '{{$cloudflareSecretName}}'
  ACME_EMAIL: '{{$acmeEmail}}'
