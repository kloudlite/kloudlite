{{- $appName :=  include "apps.messageOffice.name" . }}

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
  
  topologySpreadConstraints: {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.messageOfficeApi.replicas}}

  services:
    - port: {{ include "apps.messageOffice.httpPort" . }} 
    - port: {{ include "apps.messageOffice.privateGrpcPort" . }}
    - port: {{ include "apps.messageOffice.publicGrpcPort" . }}

  hpa:
    enabled: {{.Values.apps.messageOfficeApi.hpa.enabled}}
    minReplicas: {{.Values.apps.messageOfficeApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.messageOfficeApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.messageOfficeApi.image.repository}}:{{.Values.apps.messageOfficeApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "100m"
        max: "150m"
      resourceMemory:
        min: "100Mi"
        max: "150Mi"
      env:
        - key: HTTP_PORT
          value: {{ include "apps.messageOffice.httpPort" . | quote }}

        - key: EXTERNAL_GRPC_PORT
          value:  {{ include "apps.messageOffice.publicGrpcPort" . | quote }}

        - key: INTERNAL_GRPC_PORT
          value: {{ include "apps.messageOffice.privateGrpcPort" . | quote }}

        - key: PLATFORM_ACCESS_TOKEN
          value: {{.Values.apps.messageOfficeApi.platformAccessToken | squote}}

        - key: NATS_SEND_TO_AGENT_STREAM
          value: {{ .Values.nats.streams.sendToAgent.name }}

        - key: NATS_RECEIVE_FROM_AGENT_STREAM
          value: {{ .Values.nats.streams.receiveFromAgent.name }}

        - key: MONGO_URI
          type: secret
          refName: "mres-message-office-db-creds"
          refKey: .CLUSTER_LOCAL_URI

        - key: MONGO_DB_NAME
          type: secret
          refName: "mres-message-office-db-creds"
          refKey: DB_NAME

        - key: NATS_URL
          value: {{ include "nats.url" . | quote }}

        - key: VECTOR_GRPC_ADDR
          value: '{{ include "vector.grpc-addr" . }}'

        - key: TOKEN_HASHING_SECRET
          value: {{.Values.apps.messageOfficeApi.tokenHashingSecret | squote}}

        - key: INFRA_GRPC_ADDR
          value: '{{ include "apps.infraApi.name" . }}:{{ include "apps.infraApi.httpPort" . }}'

      livenessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.messageOffice.httpPort" . }}
        initialDelay: 5
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.messageOffice.httpPort" . }}
        initialDelay: 5
        interval: 10


