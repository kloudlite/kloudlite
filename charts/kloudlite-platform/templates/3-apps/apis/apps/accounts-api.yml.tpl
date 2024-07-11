{{- $appName := "accounts-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: accounts-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{ .Values.global.clusterSvcAccount }}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.consoleApi.configuration.replicas }}
  services:
    - port: {{.Values.apps.accountsApi.configuration.httpPort}}
    - port: {{.Values.apps.accountsApi.configuration.grpcPort}}
  containers:
    - name: main
      image: {{.Values.apps.accountsApi.image.repository}}:{{.Values.apps.accountsApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "75Mi"
        max: "100Mi"
      env:
        - key: HTTP_PORT
          value: "{{.Values.apps.accountsApi.configuration.httpPort}}"

        - key: GRPC_PORT
          value: "{{.Values.apps.accountsApi.configuration.grpcPort}}"

        - key: MONGO_URI
          type: secret
          refName: mres-accounts-db-creds
          refKey: URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-accounts-db-creds
          refKey: DB_NAME

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}

        - key: NATS_URL
          value: {{.Values.envVars.nats.url}}

        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"

        - key: IAM_GRPC_ADDR
          value: "iam:{{.Values.apps.iamApi.configuration.grpcPort}}"

        - key: COMMS_GRPC_ADDR
          value: "comms:{{.Values.apps.commsApi.configuration.grpcPort}}"

        - key: CONTAINER_REGISTRY_GRPC_ADDR
          value: "container-registry-api:{{.Values.apps.containerRegistryApi.configuration.grpcPort}}"

        - key: CONSOLE_GRPC_ADDR
          value: "console-api:{{.Values.apps.consoleApi.configuration.grpcPort}}"

        - key: AUTH_GRPC_ADDR
          value: "auth-api:{{.Values.apps.authApi.configuration.grpcPort}}"

        - key: AVAILABLE_KLOUDLITE_REGIONS_CONFIG
          value: "/kloudlite/gateways.yml"

      volumes:
        - mountPath: /kloudlite
          type: secret
          refName: {{.Values.edgeGateways.secretKeyRef.name}}
          items:
            - key: {{.Values.edgeGateways.secretKeyRef.key}}
              fileName: gateways.yml
