apiVersion: mongodb.msvc.kloudlite.io/v1
kind: {{.Values.mongo.runAsCluster | ternary "Database" "StandaloneDatabase" }}
metadata:
  name: {{.Values.envVars.db.consoleDB}}
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
    namespace: {{.Release.Namespace}}
output:
  credentialsRef:
    name: mres-{{.Values.envVars.db.consoleDB}}-creds

