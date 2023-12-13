apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.Values.managedResources.accountsDb}}
  namespace: {{.Release.Namespace}}
spec:
  resourceTemplate:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: Database

    msvcRef:
      apiVersion: mongodb.msvc.kloudlite.io/v1
      kind: ClusterService
      name: {{.Values.managedServices.mongoSvc}}

    spec:
      resourceName: {{.Values.managedResources.accountsDb}}
