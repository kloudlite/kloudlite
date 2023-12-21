apiVersion: mongodb.msvc.kloudlite.io/v1
kind: Database
metadata:
  name: {{.Values.envVars.db.infraDB}}
  namespace: {{.Release.Namespace}}
spec:
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    {{- if .Values.mongo.runAsCluster }}
    kind: ClusterService
    {{- else }}
    kind: StandaloneService
    {{- end }}
    name: mongo-svc
  resourceName: infra-db
