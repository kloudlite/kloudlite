apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "serviceAccountName" . | quote }}
  namespace: {{.Release.Namespace}}

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.Release.Namespace}}-{{- include "serviceAccountName" . }}-rb
subjects:
  - kind: ServiceAccount
    name: {{ include "serviceAccountName" . | quote }}
    namespace: {{.Release.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin

---

