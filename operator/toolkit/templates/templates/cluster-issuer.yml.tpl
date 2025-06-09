{{- $klAcmeEmail := get . "kl-acme-email" -}}
{{- $acmeDnsSolvers := get . "acme-dns-solvers" | default list -}}

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
      {{$acmeDnsSolvers |toYAML | nindent 6 }}
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
