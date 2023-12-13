apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{.Values.managedServices.mongoSvc}}
  namespace: {{.Release.Namespace}}
spec:
  serviceTemplate:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: ClusterService
    spec:
      replicas: 3
      resources:
        cpu:
          min: 400m
          max: 500m
        memory: 500Mi
        storage:
          size: 1Gi
          storageClass: {{.Values.persistence.storageClasses.xfs}}
