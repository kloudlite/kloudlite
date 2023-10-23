{{- $sa := get . "service-account-name" }} 
{{- $sa_ns := get . "namespace" }} 

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$sa}}
  namespace: {{$sa_ns}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{$sa}}-role
  namespace: {{$sa_ns}}
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get","create", "delete", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{$sa}}-rb
  namespace: {{$sa_ns}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{$sa}}-role
subjects:
  - kind: ServiceAccount
    name: {{$sa}}
    namespace: {{$sa_ns}}
---
