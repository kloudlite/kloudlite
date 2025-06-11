{{- $appName :=  include "apps.observabilityApi.name" . }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{ $appName }}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4}}
  annotations:
    {{ include "common.pod-annotations" . | nindent 4}}
spec:
  serviceAccount: {{.Values.serviceAccounts.normal.name}}

  nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 4}}
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.observabilityApi.replicas}}

  services:
    - port: 3000

  hpa:
    enabled: {{.Values.apps.observabilityApi.hpa.enabled}}
    minReplicas: {{.Values.apps.observabilityApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.observabilityApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.observabilityApi.image.repository}}:{{.Values.apps.observabilityApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "50m"
        max: "200m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"
      env:
        - key: HTTP_PORT
          value: {{ include "apps.observabilityApi.httpPort" . | quote }}

        - key: IAM_GRPC_ADDR
          value: '{{ include "apps.iamApi.name" . }}:{{ include "apps.iamApi.grpcPort" . }}'

        - key: SESSION_KV_BUCKET
          value: {{.Values.nats.buckets.sessionKV}}

        - key: NATS_URL
          value: {{ include "nats.url" . | quote }}

        - key: INFRA_GRPC_ADDR
          value: '{{ include "apps.infraApi.name" . }}:{{ include "apps.infraApi.httpPort" . }}'

        - key: ACCOUNT_COOKIE_NAME
          value: {{ include "kloudlite.account-cookie-name" . | quote }}

        - key: PROM_HTTP_ADDR
          value: {{ include "victoria-metrics.prom-url" .}}

        - key: GLOBAL_VPN_AUTHZ_SECRET
          type: secret
          refName: {{ include "apps.gatewayKubeReverseProxy.secret.name" . }}
          refKey: {{ include "apps.gatewayKubeReverseProxy.secret.key" . }}

      livenessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.observabilityApi.httpPort" . }}
        initialDelay: 5
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.observabilityApi.httpPort" . }}
        initialDelay: 5
        interval: 10

