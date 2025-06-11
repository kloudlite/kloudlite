{{- $appName := include "apps.accountsApi.name" . }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{ $appName | squote }}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4 }}
  annotations:
    kloudlite.io/checksum.edge-gateways: {{ include (print $.Template.BasePath "/03-apps/apis/secrets/edge-gateways.yml.tpl") . | sha256sum }}
    {{ include "common.pod-annotations" . | nindent 4 }}
spec:
  serviceAccount: {{.Values.serviceAccounts.normal.name}}

  nodeSelector: {{ .Values.scheduling.stateless.nodeSelector | toYaml | nindent 4 }}
  tolerations: {{ .Values.scheduling.stateless.tolerations| toYaml | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  services:
    - port: {{ include "apps.accountsApi.httpPort" . }}
    - port: {{ include "apps.accountsApi.grpcPort" . }}

  hpa:
    enabled: {{.Values.apps.accountsApi.hpa.enabled}}
    minReplicas: {{.Values.apps.accountsApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.accountsApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.accountsApi.image.repository}}:{{.Values.apps.accountsApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" . | squote}}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "75Mi"
        max: "100Mi"
      env:
        - key: HTTP_PORT
          value: {{ include "apps.accountsApi.httpPort" . | squote }}

        - key: GRPC_PORT
          value: {{ include "apps.accountsApi.grpcPort" . | squote }}

        - key: MONGO_URI
          type: secret
          refName: mres-accounts-db-creds
          refKey: .CLUSTER_LOCAL_URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-accounts-db-creds
          refKey: DB_NAME

        - key: SESSION_KV_BUCKET
          value: {{.Values.nats.buckets.sessionKVBucket.name}}

        - key: NATS_URL
          value: {{ include "nats.url" . }}

        - key: COOKIE_DOMAIN
          value: "{{- include "kloudlite.cookie-domain" . }}"

        - key: IAM_GRPC_ADDR
          value: 'iam:{{ include "apps.iamApi.grpcPort" . }}'

        - key: COMMS_GRPC_ADDR
          value: comms:{{ include "apps.commsApi.grpcPort" . }}

        - key: CONSOLE_GRPC_ADDR
          value: console-api:{{ include "apps.consoleApi.grpcPort" . }}

        - key: AUTH_GRPC_ADDR
          value: auth-api:{{ include "apps.authApi.grpcPort" . }}

        - key: AVAILABLE_KLOUDLITE_REGIONS_CONFIG
          value: "/kloudlite/gateways.yml"

      volumes:
        - mountPath: /kloudlite
          type: secret
          refName: {{ include "edge-gateways.secret.name" . }}
          items:
            - key: {{ include "edge-gateways.secret.key" . }}
              fileName: gateways.yml

      livenessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.accountsApi.httpPort" . }}
        initialDelay: 5
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.accountsApi.httpPort" . }}
        initialDelay: 5
        interval: 10

