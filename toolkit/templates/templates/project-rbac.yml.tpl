{{/* {{- $roleName := get . "role-name"  -}} */}}
{{/* {{- $roleBindingName := get . "role-binding-name"  -}} */}}

{{- $imagePullSecrets := get . "image-pull-secrets"  -}}
{{- $svcAccountName := get . "svc-account-name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $ownerRefs := get . "owner-refs" | default list  -}}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$svcAccountName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
imagePullSecrets: {{$imagePullSecrets | toYAML | nindent 2}}

{{/* --- */}}
{{/**/}}
{{/* apiVersion: rbac.authorization.k8s.io/v1 */}}
{{/* kind: Role */}}
{{/* metadata: */}}
{{/*   name: {{$roleName}} */}}
{{/*   namespace: {{$namespace}} */}}
{{/*   ownerReferences: {{$ownerRefs | toYAML | nindent 4}} */}}
{{/* rules: */}}
{{/*   - apiGroups: */}}
{{/*       - extensions */}}
{{/*       - apps */}}
{{/*     resources: */}}
{{/*       - "*" */}}
{{/*     verbs: */}}
{{/*       - "*" */}}
{{/*   - apiGroups: */}}
{{/*       - batch */}}
{{/*     resources: */}}
{{/*       - jobs */}}
{{/*       - cronjobs */}}
{{/*     verbs: */}}
{{/*       - "*" */}}
{{/**/}}
{{/* --- */}}
{{/**/}}
{{/* apiVersion: rbac.authorization.k8s.io/v1 */}}
{{/* kind:  RoleBinding */}}
{{/* metadata: */}}
{{/*   name: {{$roleName}}-rb */}}
{{/*   namespace: {{$namespace}} */}}
{{/*   ownerReferences: {{$ownerRefs | toYAML | nindent 4}} */}}
{{/* subjects: */}}
{{/*   - kind: ServiceAccount */}}
{{/*     name: {{$svcAccountName}} */}}
{{/*     namespace: {{$namespace}} */}}
{{/* roleRef: */}}
{{/*   apiGroup: rbac.authorization.k8s.io */}}
{{/*   kind: "Role" */}}
{{/*   name: {{$roleName}} */}}
