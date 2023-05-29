---
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.Values.managedResources.containerRegistryDb}}
  namespace: {{.Release.Namespace}}
spec:
  inputs:
    resourceName: {{.Values.managedResources.containerRegistryDb}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.Values.managedServices.mongoSvc}}
  mresKind:
    kind: Database
---
