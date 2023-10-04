apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Values.name}}
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.Values.name}}-rb
subjects:
  - kind: ServiceAccount
    name: {{.Values.name}}
    namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin
---
