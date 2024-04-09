{{- $sa := get . "service-account-name" }} 
{{- $sa_ns := get . "service-account-namespace" }} 

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$sa}}
  namespace: {{$sa_ns}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{$sa}}-rb
  namespace: {{$sa_ns}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: "cluster-admin"
subjects:
  - kind: ServiceAccount
    name: {{$sa}}
    namespace: {{$sa_ns}}
---
