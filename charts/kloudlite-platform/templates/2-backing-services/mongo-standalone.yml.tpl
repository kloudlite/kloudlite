apiVersion: mongodb.msvc.kloudlite.io/v1
kind: StandaloneService
metadata:
  name: mongo-svc
  namespace: {{.Release.Namespace}}
spec:
  resources:
    cpu:
      min: 300m
      max: 500m
    memory: 500Mi
    storage:
      size: 2Gi
      storageClass: sc-xfs