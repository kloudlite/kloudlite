{{- $namespace := get . "namespace" }}

{{- $sa := "kloudlite-jobs" }}

---
apiVersion: v1
kind: Namespace
metadata:
  name: {{$namespace}}
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: "{{$sa}}"
  namespace: {{$namespace}}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{$namespace}}-{{$sa}}-rb
subjects:
  - kind: ServiceAccount
    name: {{$sa}}
    namespace: {{$namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin
---
