{{- define "nodepool-operator-env" -}}
- name: CLOUD_PROVIDER_NAME
  value: {{.Values.operators.platformOperator.configuration.cloudProviderName}}

- name: CLOUD_PROVIDER_REGION
  value: {{.Values.operators.platformOperator.configuration.cloudProviderRegion}}

- name: "K3S_JOIN_TOKEN"
  value: {{.Values.operators.platformOperator.configuration.k3sJoinToken}}

- name: "K3S_SERVER_PUBLIC_HOST"
  value: {{.Values.operators.platformOperator.configuration.k3sServerPublicHost}}

- name: "KLOUDLITE_ACCOUNT_NAME"
  value: {{.Values.accountName}}

- name: KLOUDLITE_CLUSTER_NAME
  value: {{.Values.clusterName}}
{{- end}}
