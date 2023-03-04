apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{.Values.managedServices.redisSvc}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region}}
  msvcKind:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
  inputs:
    resources:
      cpu:
        min: 200m
        max: 300m
      memory: 300Mi
      storage:
        size: 1Gi
        {{/* {{if $localStorageClass}} */}}
        {{/* storageClass: {{$localStorageClass}} */}}
        {{/* {{end}} */}}
