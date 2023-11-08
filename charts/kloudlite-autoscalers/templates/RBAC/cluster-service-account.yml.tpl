{{- if .Values.serviceAccount.create }}

{{- $name := include "service-account-name" . }} 

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.Release.Namespace}}-{{$name}}-rb
subjects:
  - kind: ServiceAccount
    name: {{$name}}
    namespace: {{.Release.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin
---

{{- end }}

