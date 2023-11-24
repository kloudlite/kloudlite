{{- define "router-operator-env" -}}
- name: ACME_EMAIL
  value: {{.Values.operators.platformOperator.configuration.routers.acmeEmail}}

- name: WORKSPACE_ROUTE_SWITCHER_SERVICE
  value: "env-route-switcher"

- name: WORKSPACE_ROUTE_SWITCHER_PORT
  value: "80"
{{- end -}}
