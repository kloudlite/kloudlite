{{- $namespace := get . "namespace" -}}
{{- $acmeEmail := get . "acme-email" | default "support@kloudlite.io" -}}
{{- $defaultClusterIssuerName := get . "default-cluster-issuer-name" -}}

apiVersion: v1
kind: Secret
metadata:
  name: "router-operator-env"
  namespace: {{$namespace}}
stringData:
  ACME_EMAIL: {{$acmeEmail | squote}}
  DEFAULT_CLUSTER_ISSUER_NAME: {{$defaultClusterIssuerName}}
