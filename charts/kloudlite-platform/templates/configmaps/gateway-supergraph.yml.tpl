apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.Values.apps.gatewayApi.name}}-supergraph
  namespace: {{.Release.Namespace}}
data:
  config: |+
    serviceList:
      - name: {{.Values.apps.authApi.name}}
        url: http://{{.Values.apps.authApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query

      - name: {{.Values.apps.accountsApi.name}}
        url: http://{{.Values.apps.accountsApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query

      {{- if .Values.apps.containerRegistryApi.enabled }}
      - name: {{.Values.apps.containerRegistryApi.name}}
        url: http://{{.Values.apps.containerRegistryApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query
      {{- end }}

      - name: {{.Values.apps.consoleApi.name}}
        url: http://{{.Values.apps.consoleApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query

      {{- /* - name: {{.Values.apps.financeApi.name}} */}}
      {{- /*   url: http://{{.Values.apps.financeApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query */}}
      
      - name: {{.Values.apps.infraApi.name}}
        url: http://{{.Values.apps.infraApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query

      - name: {{.Values.apps.messageOfficeApi.name}}
        url: http://{{.Values.apps.messageOfficeApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query
---
