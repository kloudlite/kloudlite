{{- $svcAccountName := get . "svc-account-name" -}}
{{- $svcAccountNamespace := get . "svc-account-namespace" -}}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$svcAccountName}}
  namespace: {{$svcAccountNamespace}}

---
apiVersion: v1
kind: Secret
metadata:
  name: {{$svcAccountName}}
  namespace: {{$svcAccountNamespace}}
  annotations:
    kubernetes.io/service-account.name: {{$svcAccountName}}
type: kubernetes.io/service-account-token
---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{$svcAccountName}}-rb
subjects:
  - kind: ServiceAccount
    name: {{$svcAccountName}}
    namespace: {{$svcAccountNamespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
---
