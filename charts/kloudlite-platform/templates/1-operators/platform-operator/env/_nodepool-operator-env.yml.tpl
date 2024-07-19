{{- define "nodepool-operator-env" -}}
- name: CLOUD_PROVIDER_NAME
  {{- if .Values.operators.platformOperator.configuration.nodepool.extractFromCluster }}
  valueFrom:
    secretKeyRef:
      name: k3s-params
      key: cloudprovider_name
  {{- else }}
  value: {{ required ".Values.operators.platformOperator.configuration.nodepool.cloudProviderName must be set" .Values.operators.platformOperator.configuration.nodepool.cloudProviderName }} 
  {{- end }}

- name: CLOUD_PROVIDER_REGION
  {{- if .Values.operators.platformOperator.configuration.nodepool.extractFromCluster }}
  valueFrom:
    secretKeyRef:
      name: k3s-params
      key: cloudprovider_region
  {{- else }}
  value: {{ required ".Values.operators.platformOperator.configuration.nodepool.cloudProviderRegion must be set" .Values.operators.platformOperator.configuration.nodepool.cloudProviderRegion }} 
  {{- end }}

- name: "K3S_JOIN_TOKEN"
  {{- if .Values.operators.platformOperator.configuration.nodepool.extractFromCluster }}
  valueFrom:
    secretKeyRef:
      name: k3s-params
      key: k3s_agent_join_token
  {{- else }}
  value: {{ required ".Values.operators.platformOperator.configuration.nodepool.k3sAgentJoinToken must be set" .Values.operators.platformOperator.configuration.nodepool.k3sAgentJoinToken }} 
  {{- end }}

- name: "K3S_SERVER_PUBLIC_HOST"
  {{- if .Values.operators.platformOperator.configuration.nodepool.extractFromCluster }}
  valueFrom:
    secretKeyRef:
      name: k3s-params
      key: k3s_masters_public_dns_host
  {{- else }}
  value: {{ required ".Values.operators.platformOperator.configuration.nodepool.k3sServerPublicHost must be set" .Values.operators.platformOperator.configuration.nodepool.k3sServerPublicHost }}
  {{- end }}

- name: ACCOUNT_NAME
  value: {{.Values.global.accountName}}

- name: CLUSTER_NAME
  value: {{.Values.global.clusterName}}

- name: ENABLE_NODEPOOLS
  value: {{.Values.operators.platformOperator.configuration.nodepools.enabled | squote}}

{{- if .Values.operators.platformOperator.configuration.nodepools.enabled }}
- name: KLOUDLITE_RELEASE
  value: {{ .Values.apps.infraApi.configuration.kloudliteRelease | default (include "image-tag" .) }}
{{- end }}
{{- end}}
