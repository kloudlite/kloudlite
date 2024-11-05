{{- $ownerRefs := get . "owner-refs"}}
{{- $name := get . "name"}}
{{- $namespace := $name }}

{{- $dockerSecretName := get . "docker-secret-name" | default "kloudlite-docker-registry"}}
{{/*{{- $dockerConfig := get . "docker-config-json" }}*/}}

{{- $roleName := get . "role-name" | default "kloudlite-ns-admin" }}
{{- $svcAccountName := get . "svc-account-name" | default "kloudlite-svc-account" }}

{{- $accRef := get . "account-ref" }}
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{$name}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
  labels:
    kloudlite.io/account-ref: {{$accRef | squote}}
{{- if $accRef }}
---
apiVersion: artifacts.kloudlite.io/v1
kind: HarborProject
metadata:
  name: {{$accRef}}
  labels:
    kloudlite.io/account-ref: {{$accRef | squote}}
---
apiVersion: artifacts.kloudlite.io/v1
kind: HarborUserAccount
metadata:
  name: {{$dockerSecretName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  projectRef: {{$accRef}}
{{- end}}
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$svcAccountName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
---

apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{$roleName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
rules:
  - apiGroups:
      - extensions
      - apps
    resources:
      - "*"
    verbs:
      - "*"
  - apiGroups:
      - batch
    resources:
      - jobs
      - cronjobs
    verbs:
      - "*"

---

apiVersion: rbac.authorization.k8s.io/v1
kind:  RoleBinding
metadata:
  name: {{$roleName}}-rb
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
subjects:
  - kind: ServiceAccount
    name: {{$svcAccountName}}
    namespace: {{$namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "Role"
  name: {{$roleName}}

# ---
#
# apiVersion: crds.kloudlite.io/v1
# kind: AccountRouter
# metadata:
#   name: ingress-nginx
#   namespace: wg-{{$accRef}}
#   ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
# spec:
#   accountRef: {{$accRef}}
#   serviceType: ClusterIP
#   https:
#     enabled: true
#     forceRedirect: true
