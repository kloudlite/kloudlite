---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Values.clusterSvcAccount}}
  namespace: {{.Release.Namespace}}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.Release.Namespace}}-{{.Values.clusterSvcAccount}}-rb
subjects:
  - kind: ServiceAccount
    name: {{.Values.clusterSvcAccount}}
    namespace: {{.Release.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin
---
