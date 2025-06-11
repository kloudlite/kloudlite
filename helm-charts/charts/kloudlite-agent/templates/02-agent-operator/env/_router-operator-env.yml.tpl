{{- define "router-operator-env" -}}
- name: DEFAULT_CLUSTER_ISSUER
  value: {{ .Values.helmCharts.certManager.configuration.defaultClusterIssuer | quote }}

- name: DEFAULT_INGRESS_CLASS
  value: "{{.Values.helmCharts.ingressNginx.configuration.ingressClassName}}"

- name: CERTIFICATE_NAMESPACE
  value: {{.Release.Namespace}}
{{- end -}}
