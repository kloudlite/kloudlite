---
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.Values.managedResources.dnsRedis}}
  namespace: {{.Release.Namespace}}
spec:
  resourceTemplate:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: ACLAccount

    msvcRef:
      apiVersion: redis.msvc.kloudlite.io/v1
      kind: StandaloneService
      name: {{.Values.managedServices.redisSvc}}

    spec:
      keyPrefix: dns

---
