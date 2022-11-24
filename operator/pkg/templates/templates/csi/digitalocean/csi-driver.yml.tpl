{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $nodeSelector := get . "node-selector" | default dict -}}
{{/*{{- $tolerations := get . "tolerations" | default list -}}*/}}
{{- $ownerRefs := get . "owner-refs" -}}
{{- $doAccessToken := get . "do-access-token" -}}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cluster-svc-account
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML| nindent 4 }}
---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{$namespace}}-cluster-svc-account-rb
  ownerReferences: {{$ownerRefs | toYAML | nindent 4 }}
subjects:
  - kind: ServiceAccount
    name: cluster-svc-account
    namespace: {{$namespace}}
    apiGroup: ""
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: ""

---

apiVersion: csi.helm.kloudlite.io/v1
kind: DigitaloceanCSIDriver
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4 }}
spec:
  name: {{$name}}
  namespace: {{$namespace}}
  serviceAccountName: "cluster-svc-account"
  driverName: "{{$name}}"
  digitalocean:
    accessToken: {{$doAccessToken}}

  controller:
    nodeSelector: {{$nodeSelector | toYAML | nindent 6 }}
    tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists

  daemonset:
    podLabels: {}
    podAnnotations: {}
    nodeSelector: {{$nodeSelector | toYAML | nindent 6 }}
    tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
