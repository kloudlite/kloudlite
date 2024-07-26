{{- $appName := "message-office" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: message-office
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.clusterSvcAccount}}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  services:
    - port: 3000 # http
    - port: 3001 #external-grpc
    - port: 3002 #internal-grpc

  containers:
    - name: main
      image: {{.Values.apps.messageOfficeApi.image.repository}}:{{.Values.apps.messageOfficeApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      resourceCpu:
        min: "100m"
        max: "150m"
      resourceMemory:
        min: "100Mi"
        max: "150Mi"
      env:
        - key: HTTP_PORT
          value: "3000"

        - key: EXTERNAL_GRPC_PORT
          value: "3001"

        - key: INTERNAL_GRPC_PORT
          value: "3002"

        - key: PLATFORM_ACCESS_TOKEN
          value: {{.Values.apps.messageOfficeApi.configuration.platformAccessToken | squote}}

        - key: NATS_SEND_TO_AGENT_STREAM
          value: {{ .Values.envVars.nats.streams.sendToAgent.name }}

        - key: NATS_RECEIVE_FROM_AGENT_STREAM
          value: {{ .Values.envVars.nats.streams.receiveFromAgent.name }}

        - key: MONGO_URI
          type: secret
          refName: "mres-message-office-db-creds"
          refKey: .CLUSTER_LOCAL_URI

        - key: MONGO_DB_NAME
          type: secret
          refName: "mres-message-office-db-creds"
          refKey: DB_NAME

        - key: NATS_URL
          value: "nats://nats:4222"

        - key: VECTOR_GRPC_ADDR
          value: "{{.Values.vector.name}}:6000" 

        - key: TOKEN_HASHING_SECRET
          value: {{.Values.apps.messageOfficeApi.configuration.tokenHashingSecret | squote}}

        - key: INFRA_GRPC_ADDR
          value: "infra-api:3001"

      livenessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{.Values.apps.messageOfficeApi.configuration.httpPort}}
        initialDelay: 5
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{.Values.apps.messageOfficeApi.configuration.httpPort}}
        initialDelay: 5
        interval: 10


