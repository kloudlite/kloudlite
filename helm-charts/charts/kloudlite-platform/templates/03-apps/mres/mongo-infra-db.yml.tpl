apiVersion: plugin-mongodb.kloudlite.github.com/v1
kind: {{.Values.mongo.runAsCluster | ternary "Database" "StandaloneDatabase" }}
metadata:
  name: {{include "mongo.infra-db" . }}
  namespace: {{.Release.Namespace}}
spec:
  managedServiceRef:
    name: mongo-svc
    namespace: {{.Release.Namespace}}
output:
  name: mres-{{include "mongo.infra-db" . }}-creds
