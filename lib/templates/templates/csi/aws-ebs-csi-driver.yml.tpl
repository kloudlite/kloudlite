{{- $nodeSelector := get . "node-selector" -}}
{{- $tolerations := get . "tolerations" -}}
{{- $awsSecretName := get . "aws-secret-name" -}}
{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $ownerRefs := get . "owner-refs"  -}}

{{- $awsKey := get . "aws-key"  -}}
{{- $awsSecret := get . "aws-secret"  -}}

{{- $ns := printf "%s-%s" "csi" $awsSecretName -}}

---
apiVersion: v1
kind: Namespace
metadata:
  name: {{$ns}}
  ownerReferences: {{$ownerRefs | toYAML |nindent 4}}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$ns}}-cluster-svc-account
  namespace: {{$ns}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{$ns}}-cluster-svc-account-rb
subjects:
  - kind: ServiceAccount
    name: {{$ns}}-cluster-svc-account
    namespace: {{$ns}}
    apiGroup: ""
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: ""
---
apiVersion: csi.kloudlite.io/v1
kind: AwsEbsCsiDriver
metadata:
  name: {{$name}}
  namespace: csi-{{$awsSecretName}}
spec:
  fullnameOverride: {{$name}}
  controller:
    env:
      - name: AWS_ACCESS_KEY_ID
        value: {{$awsKey}}
      - name: AWS_SECRET_ACCESS_KEY
        value: {{$awsSecret}}
    serviceAccount:
      create: false
      name: {{$ns}}-cluster-svc-account
    {{- if $nodeSelector }}
    nodeSelector: {{$nodeSelector | toYAML | nindent 6}}
    {{- end }}
    {{- if $tolerations }}
    tolerations: {{$tolerations | toYAML | nindent 6}}
    {{- end}}
  node:
    tolerateAllTaints: false
    {{- if $nodeSelector }}
    nodeSelector: {{$nodeSelector | toYAML| nindent 6}}
    {{- end }}
    {{- if $tolerations }}
    tolerations: {{$tolerations | toYAML | nindent 6}}
    {{- end}}
    serviceAccount:
      create: false
      name: {{$ns}}-cluster-svc-account
---
