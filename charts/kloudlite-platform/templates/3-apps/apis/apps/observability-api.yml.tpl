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

  replicas: {{.Values.apps.observabilityApi.configuration.replicas}}

  services:
    - port: 3000

  hpa:
    enabled: true
    minReplicas: {{.Values.apps.observabilityApi.minReplicas}}
    maxReplicas: {{.Values.apps.observabilityApi.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: {{.Values.apps.observabilityApi.image.repository}}:{{.Values.apps.observabilityApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "50m"
        max: "200m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"
      env:
        - key: CLICOLOR_FORCE
          value: "1"

        - key: HTTP_PORT
          value: "3000"

        - key: IAM_GRPC_ADDR
          value: iam:3001

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}

        - key: NATS_URL
          value: {{.Values.envVars.nats.url}}

        - key: INFRA_GRPC_ADDR
          value: "infra-api:3001"

        - key: ACCOUNT_COOKIE_NAME
          value: {{.Values.global.accountCookieName}}

        - key: PROM_HTTP_ADDR
          value: {{ include "prom-http-addr" .}}

        - key: GLOBAL_VPN_AUTHZ_SECRET
          value: {{.Values.apps.infraApi.configuration.globalVpnKubeReverseProxyAuthzToken}}

      livenessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: 3000
        initialDelay: 5
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: 3000
        initialDelay: 5
        interval: 10

