{{- $nodeSelector := get . "node-selector" -}}
{{- $awsSecretName := get . "aws-secret-name" -}}
{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $ownerRefs := get . "owner-refs"  -}}

{{- $awsKey := get . "aws-key"  -}}
{{- $awsSecret := get . "aws-secret"  -}}
{{- $svcAccountName := get . "svc-account-name"  -}}

{{/*{{- $ns := printf "%s-%s" "csi" $awsSecretName -}}*/}}

{{/*---*/}}
{{/*apiVersion: v1*/}}
{{/*kind: Namespace*/}}
{{/*metadata:*/}}
{{/*  name: {{$ns}}*/}}
{{/*  ownerReferences: {{$ownerRefs | toYAML |nindent 4}}*/}}
{{/*---*/}}
{{/*apiVersion: v1*/}}
{{/*kind: ServiceAccount*/}}
{{/*metadata:*/}}
{{/*  name: {{$ns}}-cluster-svc-account*/}}
{{/*  namespace: {{$ns}}*/}}
{{/*---*/}}
{{/*apiVersion: rbac.authorization.k8s.io/v1*/}}
{{/*kind: ClusterRoleBinding*/}}
{{/*metadata:*/}}
{{/*  name: {{$ns}}-cluster-svc-account-rb*/}}
{{/*subjects:*/}}
{{/*  - kind: ServiceAccount*/}}
{{/*    name: {{$ns}}-cluster-svc-account*/}}
{{/*    namespace: {{$ns}}*/}}
{{/*    apiGroup: ""*/}}
{{/*roleRef:*/}}
{{/*  kind: ClusterRole*/}}
{{/*  name: cluster-admin*/}}
{{/*  apiGroup: ""*/}}
{{/*---*/}}
apiVersion: "storage.k8s.io/v1"
kind: CSIDriver
metadata:
  name: ebs.csi.aws.com
spec:
  attachRequired: true
  podInfoOnMount: false
---
apiVersion: csi.helm.kloudlite.io/v1
kind: AwsEbsCsiDriver
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML |nindent 4}}
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
      name: {{$svcAccountName}}
    {{- if $nodeSelector }}
    nodeSelector: {{$nodeSelector | toYAML | nindent 6}}
    {{- end }}
    tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
  node:
    tolerateAllTaints: false
    {{- if $nodeSelector }}
    nodeSelector: {{$nodeSelector | toYAML| nindent 6}}
    {{- end }}
    tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
    serviceAccount:
      create: false
      name: {{$svcAccountName}}
---
