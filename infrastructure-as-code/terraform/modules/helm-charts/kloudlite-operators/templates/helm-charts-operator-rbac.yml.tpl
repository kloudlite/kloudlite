apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${svc_account_name}
  namespace: ${svc_account_namespace}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${svc_account_name}-rb
subjects:
  - kind: ServiceAccount
    name: ${svc_account_name}
    namespace: ${svc_account_namespace}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin
