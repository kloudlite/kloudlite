{{- if .Values.serviceAccounts.clusterAdmin.create }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Values.serviceAccounts.clusterAdmin.name}}
  namespace: {{.Release.Namespace}}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.Release.Namespace}}-{{.Values.serviceAccounts.clusterAdmin.name}}-rb
subjects:
  - kind: ServiceAccount
    name: {{.Values.serviceAccounts.clusterAdmin.name}}
    namespace: {{.Release.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin
---
{{- end }}
