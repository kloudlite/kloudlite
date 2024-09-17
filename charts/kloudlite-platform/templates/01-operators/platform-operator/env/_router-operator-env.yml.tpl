{{- define "router-operator-env" -}}
- name: DEFAULT_CLUSTER_ISSUER
  value: {{.Values.certManager.clusterIssuer.name}}

- name: DEFAULT_INGRESS_CLASS
  value: {{.Values.operators.platformOperator.configuration.ingressClassName}}

- name: CERTIFICATE_NAMESPACE
  value: {{.Release.Namespace}}
{{- end -}}
