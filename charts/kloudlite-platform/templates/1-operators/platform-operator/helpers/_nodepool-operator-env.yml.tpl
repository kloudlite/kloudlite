{{- define "nodepool-operator-env" -}}
- name: CLOUD_PROVIDER_NAME
  value: {{ required ".Values.operators.platformOperator.configuration.nodepool.cloudProviderName must be set" .Values.operators.platformOperator.configuration.nodepool.cloudProviderName }} 
- name: CLOUD_PROVIDER_REGION
  value: {{ required ".Values.operators.platformOperator.configuration.nodepool.cloudProviderRegion must be set" .Values.operators.platformOperator.configuration.nodepool.cloudProviderRegion }} 

- name: "K3S_JOIN_TOKEN"
  value: {{ required ".Values.operators.platformOperator.configuration.nodepool.k3sJoinToken must be set" .Values.operators.platformOperator.configuration.nodepool.k3sJoinToken }} 

- name: "K3S_SERVER_PUBLIC_HOST"
  value: {{ required ".Values.operators.platformOperator.configuration.nodepool.k3sServerPublicHost must be set" .Values.operators.platformOperator.configuration.nodepool.k3sServerPublicHost }} 

- name: "KLOUDLITE_ACCOUNT_NAME"
  value: {{.Values.global.accountName}}

- name: KLOUDLITE_CLUSTER_NAME
  value: {{.Values.global.clusterName}}
{{- end}}
