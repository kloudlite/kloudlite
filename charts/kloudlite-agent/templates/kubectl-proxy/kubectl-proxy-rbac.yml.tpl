{{- $name := printf "%s-kubectl-proxy" .Release.Name }}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{$name}}-role
  namespace: {{.Release.Namespace}}
rules:
- apiGroups: [""]
  resources: ["pods", "pods/log", "namespaces"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{$name}}-rb
  namespace: {{.Release.Namespace}}
subjects:
  - kind: ServiceAccount
    name: {{$name}}
    namespace: {{.Release.Namespace}}
roleRef:
  kind: ClusterRole
  name: {{$name}}-role
  apiGroup: "rbac.authorization.k8s.io"
---
