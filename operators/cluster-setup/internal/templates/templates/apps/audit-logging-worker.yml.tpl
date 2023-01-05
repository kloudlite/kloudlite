{{- $namespace := get . "namespace" -}}
{{- $svcAccount := get . "svc-account" -}}
{{- $sharedConstants := get . "shared-constants" -}}

{{- $ownerRefs := get . "owner-refs" | default list -}}
{{- $accountRef := get . "account-ref" | default "kl-core" -}}
{{- $region := get . "region" | default "master" -}}
{{- $imagePullPolicy := get . "image-pull-policy" | default "Always" -}}

{{- $nodeSelector := get . "node-selector" | default dict -}}
{{- $tolerations := get . "tolerations" | default list -}}

{{ with $sharedConstants}}
{{/*gotype: operators.kloudlite.io/apis/cluster-setup/v1.SharedConstants*/}}
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.EventsDbName}}
  namespace: {{$namespace}}
  labels:
    kloudlite.io/account-ref:  {{$accountRef}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  inputs:
    resourceName: {{.EventsDbName}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.MongoSvcName}}
  mresKind:
    kind: Database
{{end}}
