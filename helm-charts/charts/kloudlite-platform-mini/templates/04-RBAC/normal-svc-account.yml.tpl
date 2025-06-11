{{- if .Values.serviceAccounts.normal.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Values.serviceAccounts.normal.name}}
  namespace: {{.Release.Namespace}}
{{- end }}
