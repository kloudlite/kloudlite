apiVersion: plugin-mongodb.kloudlite.github.com/v1
kind: {{.Values.mongo.runAsCluster | ternary "Database" "StandaloneDatabase" }}
metadata:
  name: {{include "mongo.message-office-db" . }}
  namespace: {{.Release.Namespace}}
spec:
  managedServiceRef:
    name: mongo-svc
    namespace: {{.Release.Namespace}}
output:
  name: mres-{{include "mongo.message-office-db" . }}-creds
