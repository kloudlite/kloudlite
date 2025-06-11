{{- $appName := include "apps.iamApi.name" . }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{$appName}}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4}}
  annotations:
    {{ include "common.pod-annotations" . | nindent 4}}
spec:
  serviceAccount: {{.Values.serviceAccounts.normal.name}}
  nodeSelector: {{ .Values.scheduling.stateless.nodeSelector| toYaml | nindent 4 }}
  tolerations: {{ .Values.scheduling.stateless.tolerations | toYaml | nindent 4 }}

  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.iamApi.replicas}}

  services:
    - port: {{ include "apps.iamApi.grpcPort" . }}

  hpa:
    enabled: {{.Values.apps.iamApi.hpa.enabled}}
    minReplicas: {{.Values.apps.iamApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.iamApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.iamApi.image.repository}}:{{.Values.apps.iamApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      
      resourceCpu:
        min: "30m"
        max: "50m"
      resourceMemory:
        min: "50Mi"
        max: "100Mi"
      env:
        - key: MONGO_DB_URI
          type: secret
          refName: mres-iam-db-creds
          refKey: .CLUSTER_LOCAL_URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-iam-db-creds
          refKey: DB_NAME

        - key: COOKIE_DOMAIN
          value: "{{- include "kloudlite.cookie-domain" . }}"

        - key: ACCOUNTS_GRPC_ADDR
          value: '{{ include "apps.accountsApi.name" . }}:{{ include "apps.accountsApi.grpcPort" . }}'

        - key: GRPC_PORT
          value: {{ include "apps.iamApi.grpcPort" . | squote }}

        - key: CONSOLE_SERVICE
          value: {{ include "apps.consoleApi.name" . }}:{{ include "apps.consoleApi.grpcPort" . }}
