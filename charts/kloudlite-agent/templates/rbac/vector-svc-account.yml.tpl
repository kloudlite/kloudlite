{{/* INFO: Vector Svc Account is required, as we are running kubelet-metrics-reexporter as a sidecar in vector pod. This sidecar needs to access kubelet metrics and hence we need to create a service account with required permissions. */}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ .Values.vectorSvcAccountName }}
  namespace: {{ .Release.Namespace }}
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  - nodes
  - pods
  verbs:
  - list
  - watch

- apiGroups:
  - ""
  resources:
  - nodes/proxy
  verbs:
  - get

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{ .Values.vectorSvcAccountName }}-rb
subjects:
  - kind: ServiceAccount
    name: {{.Values.vectorSvcAccountName}}
    namespace: {{.Release.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ .Values.vectorSvcAccountName }}

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .Values.vectorSvcAccountName }}
  namespace: {{.Release.Namespace}}

