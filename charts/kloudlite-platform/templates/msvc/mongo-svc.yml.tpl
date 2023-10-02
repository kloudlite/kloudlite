apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{.Values.managedServices.mongoSvc}}
  namespace: {{.Release.Namespace}}
spec:
  nodeSelector: {{.Values.managedServicesNodeSelector |toYaml | nindent 6 }}
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
        storageClass: {{.Values.persistence.storageClasses.xfs}}
