---
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.Values.managedResources.containerRegistryRedis}}
  namespace: {{.Release.Namespace}}
spec:
  inputs:
    keyPrefix: container-registry
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.Values.managedServices.redisSvc}}

---
