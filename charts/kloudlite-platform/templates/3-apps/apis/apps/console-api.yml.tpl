{{- $appName := "console-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: console-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.clusterSvcAccount}}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.consoleApi.configuration.replicas}}

  services:
    - port: {{.Values.apps.consoleApi.configuration.httpPort | int }}
    - port: {{.Values.apps.consoleApi.configuration.grpcPort | int }}

  containers:
    - name: main
      image: {{.Values.apps.consoleApi.image.repository}}:{{.Values.apps.consoleApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
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
          value: {{.Values.apps.consoleApi.configuration.httpPort | squote}}

        - key: GRPC_PORT
          value: {{.Values.apps.consoleApi.configuration.grpcPort | squote}}

        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"

        - key: DNS_ADDR
          value: :5353

        - key: KLOUDLITE_DNS_SUFFIX
          value: "{{.Values.global.kloudliteDNSSuffix}}"

        - key: MONGO_URI
          type: secret
          refName: mres-console-db-creds
          refKey: .CLUSTER_LOCAL_URI

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

        - key: NATS_RECEIVE_FROM_AGENT_STREAM
          value: {{.Values.envVars.nats.streams.receiveFromAgent.name}}

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
          value: {{include "prom-http-addr" .}}

        - key: DEVICE_NAMESPACE
          value: {{.Values.apps.consoleApi.configuration.vpnDeviceNamespace}}

      volumes:
        - mountPath: /console.d/templates
          type: config
          refName: managed-svc-template
          items:
            - key: managed-svc-templates.yml

      livenessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{.Values.apps.consoleApi.configuration.httpPort}}
        initialDelay: 5
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{.Values.apps.consoleApi.configuration.httpPort}}
        initialDelay: 5
        interval: 10
