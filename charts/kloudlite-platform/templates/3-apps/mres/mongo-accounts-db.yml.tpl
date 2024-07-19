apiVersion: mongodb.msvc.kloudlite.io/v1
kind: {{.Values.mongo.runAsCluster | ternary "Database" "StandaloneDatabase" }}
metadata:
  name: {{.Values.envVars.db.accountsDB}}
  namespace: {{.Release.Namespace}}
spec:
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: {{.Values.mongo.runAsCluster | ternary "ClusterService" "StandaloneService" }}
    name: mongo-svc
    namespace: {{.Release.Namespace}}
output:
  credentialsRef:
    name: mres-{{.Values.envVars.db.accountsDB}}-creds
