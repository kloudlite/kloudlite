{{- $namespace := get . "namespace" -}}
{{- $region := get . "region" -}}
{{- $localStorageClass := get . "local-storage-class" -}}
{{- $sharedConstants := get . "shared-constants" -}}
{{- $ownerRefs := get . "owner-refs" | default list -}}

{{- with $sharedConstants -}}
{{/*gotype: operators.kloudlite.io/apis/cluster-setup/v1.SharedConstants*/}}
apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{.MongoSvcName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  region: {{$region}}
  msvcKind:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
  inputs:
    resources:
      cpu:
        min: 300m
        max: 500m
      memory: 400Mi
      storage:
        size: 1Gi
        {{ if $localStorageClass }}
        storageClass: {{$localStorageClass}}
        {{end}}
---
# auth-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.AuthDbName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  inputs:
    resourceName: {{.AuthDbName}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.MongoSvcName}}
  mresKind:
    kind: Database

---
# ci-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.CiDbName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  inputs:
    resourceName: {{.CiDbName}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.MongoSvcName}}
  mresKind:
    kind: Database

---
# console-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.ConsoleDbName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  inputs:
    resourceName: {{.ConsoleDbName}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.MongoSvcName}}
  mresKind:
    kind: Database
---
# dns-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.DnsDbName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  inputs:
    resourceName: {{.DnsDbName}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.MongoSvcName}}
  mresKind:
    kind: Database
---
# finance-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.FinanceDbName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  inputs:
    resourceName: {{.FinanceDbName}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.MongoSvcName}}
  mresKind:
    kind: Database
---
# iam db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.IamDbName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  inputs:
    resourceName: {{.IamDbName}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.MongoSvcName}}
  mresKind:
    kind: Database

---
# comms db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: comms-db
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  inputs:
    resourceName: {{.CommsDbName}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.MongoSvcName}}
  mresKind:
    kind: Database
{{end}}
