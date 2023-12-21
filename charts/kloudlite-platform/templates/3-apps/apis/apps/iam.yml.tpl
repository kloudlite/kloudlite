apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: iam
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.iamApi.image}}
      imagePullPolicy: {{.Values.global.imagePullPolicy }}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      
      resourceCpu:
        min: "30m"
        max: "50m"
      resourceMemory:
        min: "50Mi"
        max: "100Mi"
      env:

        - key: MONGO_DB_URI
          type: secret
          refName: mres-iam-db-creds
          refKey: URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-iam-db-creds
          refKey: DB_NAME

        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"

        - key: GRPC_PORT
          value: "3001"

        - key: CONSOLE_SERVICE
          value: "console-api:3001"
