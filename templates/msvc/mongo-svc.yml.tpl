apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{.Values.managedServices.mongoSvc}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region}}
  msvcKind:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
  inputs:
    resources:
      cpu:
        min: 400m
        max: 500m
      memory: 500Mi
      storage:
        size: 1Gi
        {{/* {{ if $localStorageClass }} */}}
        {{/* storageClass: {{$localStorageClass}} */}}
        {{/* {{end}} */}}
---
{{/* # comms db */}}
{{/* apiVersion: crds.kloudlite.io/v1 */}}
{{/* kind: ManagedResource */}}
{{/* metadata: */}}
{{/*   name: comms-db */}}
{{/*   namespace: {{$namespace}} */}}
{{/*   ownerReferences: {{$ownerRefs| toYAML | nindent 4}} */}}
{{/* spec: */}}
{{/*   inputs: */}}
{{/*     resourceName: {{.CommsDbName}} */}}
{{/*   msvcRef: */}}
{{/*     apiVersion: mongodb.msvc.kloudlite.io/v1 */}}
{{/*     kind: StandaloneService */}}
{{/*     name: {{.MongoSvcName}} */}}
{{/*   mresKind: */}}
{{/*     kind: Database */}}
