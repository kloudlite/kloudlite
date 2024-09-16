{{- $appName := "console-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{$appName}}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4 }}
  annotations: {{ include "common.pod-annotations" . | nindent 4 }}
spec:
  serviceAccount: {{.Values.serviceAccounts.clusterAdmin.name}}

  nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 4}}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.consoleApi.replicas}}

  services:
    - port: {{ include "apps.consoleApi.httpPort" . }}
    - port: {{ include "apps.consoleApi.grpcPort" . }}

  hpa:
    enabled: {{.Values.apps.consoleApi.hpa.enabled}}
    minReplicas: {{.Values.apps.consoleApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.consoleApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.consoleApi.image.repository}}:{{.Values.apps.consoleApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "80m"
        max: "200m"
      resourceMemory:
        min: "80Mi"
        max: "150Mi"
      env:
        - key: HTTP_PORT
          value: {{ include "apps.consoleApi.httpPort" . | squote}}

        - key: GRPC_PORT
          value: {{ include "apps.consoleApi.grpcPort" . | squote}}

        - key: COOKIE_DOMAIN
          value: "{{- include "kloudlite.cookie-domain" . }}"

        - key: DNS_ADDR
          value: ':{{ include "apps.consoleApi.dnsPort" . }}'

        - key: KLOUDLITE_DNS_SUFFIX
          value: {{required ".Values.kloudliteDNSSuffix is required" .Values.kloudliteDNSSuffix | squote}}

        - key: MONGO_URI
          type: secret
          refName: mres-console-db-creds
          refKey: .CLUSTER_LOCAL_URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-console-db-creds
          refKey: DB_NAME

        - key: ACCOUNT_COOKIE_NAME
          value: {{ include "kloudlite.account-cookie-name" . | squote }}

        - key: CLUSTER_COOKIE_NAME
          value: {{ include "kloudlite.account-cookie-name" . | squote }}

        - key: NATS_URL
          value: {{ include "nats.url" . }}

        - key: NATS_RECEIVE_FROM_AGENT_STREAM
          value: {{.Values.nats.streams.receiveFromAgent.name}}

        - key: EVENTS_NATS_STREAM
          value: {{.Values.nats.streams.events.name}}

        - key: SESSION_KV_BUCKET
          value: {{.Values.nats.buckets.sessionKVBucket.name}}

        - key: IAM_GRPC_ADDR
          value: iam:{{ include "apps.iamApi.grpcPort" . }}

        - key: INFRA_GRPC_ADDR
          value: infra-api:{{ include "apps.infraApi.grpcPort" . }}

        - key: ACCOUNT_GRPC_ADDR
          value: accounts-api:{{ include "apps.accountsApi.grpcPort" . }}

        - key: MESSAGE_OFFICE_INTERNAL_GRPC_ADDR
          value: message-office:{{ include "apps.messageOffice.privateGrpcPort" . }}

        - key: CONSOLE_CACHE_KV_BUCKET
          value: {{.Values.nats.buckets.consoleCacheBucket.name}}

        - key: MSVC_TEMPLATE_FILE_PATH
          value: /console.d/templates/managed-svc-templates.yml

        - key: WEBHOOK_TOKEN_HASHING_SECRET
          {{- /* FIXME: this is a secret, that should be generated */}}
          value: {{.Values.apps.webhooksApi.webhookAuthzTokenHashingSecret | squote}}

        - key: WEBHOOK_URL
          value: "https://webhooks.{{.Values.baseDomain}}"

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
          port: {{ include "apps.consoleApi.httpPort" . }}
        initialDelay: 10
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.consoleApi.httpPort" . }}
        initialDelay: 10
        interval: 10
