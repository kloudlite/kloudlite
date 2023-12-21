apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: mongo-svc
  namespace: {{.Release.Namespace}}
spec:
  serviceTemplate:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    {{- if .Values.mongo.runAsCluster }}
    kind: ClusterService
    {{- else }}
    kind: StandaloneService
    {{- end }}
    spec:
    {{- if .Values.mongo.runAsCluster }}
      replicas: {{.Values.mongo.replicas}}
    {{- end }}
      resources:
        cpu:
          min: 400m
          max: 500m
        memory: 500Mi
        storage:
          size: 1Gi
          storageClass: {{.Values.persistence.storageClasses.xfs}}
