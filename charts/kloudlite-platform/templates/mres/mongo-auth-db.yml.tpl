apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.Values.managedResources.authDb}}
  namespace: {{.Release.Namespace}}
  labels:
    
spec:
  inputs:
    resourceName: {{.Values.managedResources.authDb}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.Values.managedServices.mongoSvc}}
  mresKind:
    kind: Database
