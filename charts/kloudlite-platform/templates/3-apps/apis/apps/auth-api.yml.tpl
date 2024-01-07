apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: auth-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{ .Values.global.clusterSvcAccount }}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.authApi.image}}
      imagePullPolicy: {{.Values.global.imagePullPolicy }}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      resourceCpu:
        min: "50m"
        max: "60m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"
      env:

        - key: MONGO_URI
          type: secret
          refName: mres-auth-db-creds
          refKey: URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-auth-db-creds
          refKey: DB_NAME

        - key: COMMS_SERVICE
          value: "comms:3001"

        - key: PORT
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}

        - key: RESET_PASSWORD_TOKEN_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.resetTokenBucket.name}}

        - key: VERIFY_TOKEN_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.verifyTokenBucketName}}

        - key: NATS_URL
          value: {{.Values.envVars.nats.url}}

        - key: ORIGINS
          {{/* value: "https://{{.AuthWebDomain}},http://localhost:4001,https://studio.apollographql.com" */}}
          value: "https://kloudlite.io,http://localhost:4001,https://studio.apollographql.com"

        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"


        {{- if .Values.oAuth.providers.github.enabled }}
        - key: GITHUB_APP_PK_FILE
          value: /github/github-app-pk.pem
        {{- end }}

      envFrom:
        - type: secret
          refName: {{.Values.oAuth.secretName}}

      {{- if .Values.oAuth.providers.github.enabled }}
      volumes:
        - mountPath: /github
          type: secret
          refName: {{.Values.oAuth.secretName}}
          items:
            - key: github-app-pk.pem
              fileName: github-app-pk.pem
      {{- end }}
