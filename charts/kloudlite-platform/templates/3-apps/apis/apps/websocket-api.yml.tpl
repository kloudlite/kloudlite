{{- $appName := "websocket-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: websocket-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{ .Values.global.clusterSvcAccount }}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.websocketApi.configuration.replicas}}

  services:
    - port: 3000

  containers:
    - name: main
      image: {{.Values.apps.websocketApi.image.repository}}:{{.Values.apps.websocketApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"
      env:
        - key: SOCKET_PORT
          value: "3000"

        - key: IAM_GRPC_ADDR
          value: iam:3001
        
        - key: OBSERVABILITY_API_ADDR
          value: observability-api:3000

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}

        - key: NATS_URL
          value: {{.Values.envVars.nats.url}}

        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"
