{{- $accRef := get . "acc-ref"  -}}
{{- $dockerSecretName := get . "docker-secret-name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $ownerRefs := get . "owner-refs"  -}}
{{- $projectName := get . "project-name"  -}}

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
  name: {{$projectName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  projectRef: {{$accRef}}
