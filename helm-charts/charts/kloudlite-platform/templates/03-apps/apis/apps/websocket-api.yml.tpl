{{- $appName := include "apps.websocketApi.name" . -}}

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

  replicas: {{.Values.apps.websocketApi.replicas}}

  services:
    - port: {{ include "apps.websocketApi.httpPort" . }}

  hpa:
    enabled: {{.Values.apps.websocketApi.hpa.enabled}}
    minReplicas: {{.Values.apps.websocketApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.websocketApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.websocketApi.image.repository}}:{{.Values.apps.websocketApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"
      env:
        - key: SOCKET_PORT
          value: {{ include "apps.websocketApi.httpPort" . | squote }}

        - key: IAM_GRPC_ADDR
          value: 'iam:{{ include "apps.iamApi.grpcPort" . }}'
        
        - key: OBSERVABILITY_API_ADDR
          value: '{{ include "apps.observabilityApi.name" . }}:{{ include "apps.observabilityApi.httpPort" . }}'

        - key: SESSION_KV_BUCKET
          value: {{.Values.nats.buckets.sessionKVBucket.name}}

        - key: NATS_URL
          value: {{ include "nats.url" . | quote}}

        - key: COOKIE_DOMAIN
          value: {{ include "kloudlite.cookie-domain" . }}
