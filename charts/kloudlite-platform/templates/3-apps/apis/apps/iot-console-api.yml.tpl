{{- $appName := "iot-console-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{$appName}}
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.iotConsoleApi.configuration.replicas}}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.iotConsoleApi.image.repository}}:{{.Values.apps.iotConsoleApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
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
        - key: HTTP_PORT
          value: "3000"

        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"

        - key: ACCOUNT_COOKIE_NAME
          value: {{.Values.global.accountCookieName}}

        - key: CLUSTER_COOKIE_NAME
          value: {{.Values.global.clusterCookieName}}

        - key: MONGO_URI
          type: secret
          refName: "mres-{{.Values.envVars.db.iotConsoleDB}}-creds"
          refKey: URI

        - key: MONGO_DB_NAME
          type: secret
          refName: "mres-{{.Values.envVars.db.iotConsoleDB}}-creds"
          refKey: DB_NAME

        - key: NATS_URL
          value: {{.Values.envVars.nats.url}}

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}

        - key: IOT_CONSOLE_CACHE_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.iotConsoleCacheBucket.name}}

