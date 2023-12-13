apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.Values.managedResources.authDb}}
  namespace: {{.Release.Namespace}}
  labels:
    
spec:
  resourceTemplate:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: Database

    msvcRef:
      apiVersion: mongodb.msvc.kloudlite.io/v1
      kind: ClusterService
      name: {{.Values.managedServices.mongoSvc}}

    spec:
      resourceName: {{.Values.managedResources.authDb}}
