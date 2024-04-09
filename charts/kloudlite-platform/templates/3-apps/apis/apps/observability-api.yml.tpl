{{- $appName := "observability-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: observability-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{ .Values.global.clusterSvcAccount }}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.observabilityApi.configuration.replicas}}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.observabilityApi.image.repository}}:{{.Values.apps.observabilityApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"
      env:
        - key: HTTP_PORT
          value: "3000"

        - key: IAM_GRPC_ADDR
          value: iam:3001

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}

        - key: NATS_URL
          value: {{.Values.envVars.nats.url}}

        - key: ACCOUNT_COOKIE_NAME
          value: {{.Values.global.accountCookieName}}

        - key: PROM_HTTP_ADDR
          value: {{ include "prom-http-addr" .}}
