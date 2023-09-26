apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{.Values.managedServices.redisSvc}}
  namespace: {{.Release.Namespace}}
spec:
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
        storageClass: {{.Values.persistence.storageClasses.ext4}}
