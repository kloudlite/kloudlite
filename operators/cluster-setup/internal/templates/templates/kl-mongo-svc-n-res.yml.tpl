{{/*{{- $namespace := get . "namespace" -}}*/}}
{{- $namespace := get . "namespace" -}}
{{- $region := get . "region" -}}
{{- $localStorageClass := get . "local-storage-class" -}}

{{- $mongoSvcName := "mongo-svc" -}}
apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{$mongoSvcName}}
  namespace: {{$namespace}}
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
      storageClass: $localStorageClass
      {{end}}
---
# auth-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: auth-db
  namespace: {{$namespace}}
spec:
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$mongoSvcName}}
  mresKind:
    kind: Database

---
# ci-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: ci-db
  namespace: {{$namespace}}
spec:
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$mongoSvcName}}
  mresKind:
    kind: Database

---
# console-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: console-db
  namespace: {{$namespace}}
spec:
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$mongoSvcName}}
  mresKind:
    kind: Database
---
# dns-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: dns-db
  namespace: {{$namespace}}
spec:
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$mongoSvcName}}
  mresKind:
    kind: Database
---
# finance-db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: finance-db
  namespace: {{$namespace}}
spec:
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$mongoSvcName}}
  mresKind:
    kind: Database
---
# iam db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: iam-db
  namespace: {{$namespace}}
spec:
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$mongoSvcName}}
  mresKind:
    kind: Database

---
# iam db
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: comms-db
  namespace: {{$namespace}}
spec:
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$mongoSvcName}}
  mresKind:
    kind: Database
