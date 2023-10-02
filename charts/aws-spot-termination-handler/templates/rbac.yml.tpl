apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-spot-k3s-termination-handler
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aws-spot-k3s-termination-handler-rb
subjects:
  - kind: ServiceAccount
    name: aws-spot-k3s-termination-handler
    namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin
---
