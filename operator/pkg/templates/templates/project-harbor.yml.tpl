{{- $accRef := get . "acc-ref"  -}}
{{- $dockerSecretName := get . "docker-secret-name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $ownerRefs := get . "owner-refs"  -}}

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
