apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: console-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.clusterSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}
  
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp

    - port: {{.Values.apps.consoleApi.configuration.logsAndMetricsHttpPort | int }}
      targetPort: {{.Values.apps.consoleApi.configuration.logsAndMetricsHttpPort | int }}
      name: http
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.consoleApi.image}}
      imagePullPolicy: {{.Values.global.imagePullPolicy }}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      resourceCpu:
        min: "80m"
        max: "150m"
      resourceMemory:
        min: "80Mi"
        max: "150Mi"
      env:
        - key: HTTP_PORT
          value: "3000"

        - key: LOGS_AND_METRICS_HTTP_PORT
          value: {{.Values.apps.consoleApi.configuration.logsAndMetricsHttpPort | squote}}
          {{- /* LOGS_AND_METRICS_HTTP_PORT=9999 */}}

        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"

        - key: MONGO_URI
          type: secret
          refName: mres-console-db-creds
          refKey: URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-console-db-creds
          refKey: DB_NAME

        - key: ACCOUNT_COOKIE_NAME
          value: {{.Values.global.accountCookieName}}

        - key: CLUSTER_COOKIE_NAME
          value: {{.Values.global.clusterCookieName}}

        - key: NATS_URL
          value: {{.Values.envVars.nats.url}}

        - key: NATS_RESOURCE_STREAM
          value: {{.Values.envVars.nats.streams.resourceSync.name}}

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}

        - key: IAM_GRPC_ADDR
          value: "iam:3001"

        - key: INFRA_GRPC_ADDR
          value: "infra-api:3001"

        - key: DEFAULT_PROJECT_WORKSPACE_NAME
          value: {{.Values.global.defaultProjectWorkspaceName}}

        - key: CONSOLE_CACHE_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.consoleCacheBucket.name}}

        - key: MSVC_TEMPLATE_FILE_PATH
          value: /console.d/templates/managed-svc-templates.yml

        - key: LOKI_SERVER_HTTP_ADDR
          value: http://{{ .Values.loki.name }}.{{.Release.Namespace}}.svc.{{.Values.global.clusterInternalDNS}}:3100

        - key: PROM_HTTP_ADDR
          value: http://{{ .Values.prometheus.name }}-prometheus.{{.Release.Namespace}}.svc.{{.Values.global.clusterInternalDNS}}:9090

      volumes:
        - mountPath: /console.d/templates
          type: config
          refName: managed-svc-template
          items:
            - key: managed-svc-templates.yml