{{- define "router-operator-env" -}}

- name: WORKSPACE_ROUTE_SWITCHER_SERVICE
  value: "env-route-switcher"

- name: WORKSPACE_ROUTE_SWITCHER_PORT
  value: "80"

- name: DEFAULT_CLUSTER_ISSUER
  value: {{.Values.certManager.certIssuer.name}}

- name: DEFAULT_INGRESS_CLASS
  value: {{.Values.global.ingressClassName}}

- name: CERTIFICATE_NAMESPACE
  value: {{.Release.Namespace}}
{{- end -}}
