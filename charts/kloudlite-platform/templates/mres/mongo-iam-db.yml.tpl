---
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.Values.managedResources.iamDb}}
  namespace: {{.Release.Namespace}}
  labels:
    
spec:
  inputs:
    resourceName: {{.Values.managedResources.iamDb}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.Values.managedServices.mongoSvc}}
  mresKind:
    kind: Database
---
