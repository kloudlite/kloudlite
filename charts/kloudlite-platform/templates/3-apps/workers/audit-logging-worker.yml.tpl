apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: audit-logging
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}
  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services: []
  containers:
    - name: main
      image: {{.Values.apps.auditLoggingWorker.image}}
      imagePullPolicy: {{.Values.global.imagePullPolicy}}
      resourceCpu:
        min: "50m"
        max: "70m"
      resourceMemory:
        min: "50Mi"
        max: "70Mi"
      env:
        - key: DB_URI
          type: secret
          refName: mres-events-db-creds
          refKey: URI
        - key: DB_NAME
          type: secret
          refName: mres-events-db-creds
          refKey: DB_NAME
        - key: NATS_URL
          value: {{.Values.envVars.nats.url}}
        - key: EVENT_LOG_NATS_STREAM
          value: {{.Values.envVars.nats.streams.events.name}}
