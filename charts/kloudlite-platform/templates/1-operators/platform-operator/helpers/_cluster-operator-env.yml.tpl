{{- $operatorName := "kloudlite-platform-operator" }}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{$operatorName}}-cloudflare-params
  namespace: {{.Release.Namespace}}
data:
  api_token: {{.Values.operators.platformOperator.configuration.cluster.cloudflare.apiToken | b64enc | quote }}
  base_domain: {{.Values.operators.platformOperator.configuration.cluster.cloudflare.baseDomain | b64enc | quote }}
  zone_id: {{.Values.operators.platformOperator.configuration.cluster.cloudflare.zoneId | b64enc | quote }}
---

{{- define "cluster-operator-env" -}}
{{- $operatorName := "kloudlite-platform-operator" }}
{{- $clusterOperatorVars := .Values.operators.platformOperator.configuration.cluster }} 

- name: CLOUDFLARE_API_TOKEN
  valueFrom:
    secretKeyRef:
      name: kloudlite-cloudflare
      key: api_token

- name: NATS_URL
  value: nats://nats:4222

- name: CLOUDFLARE_ZONE_ID
  value: {{ required ".Values.operators.platformOperator.configuration.cluster.cloudflare.zoneId must be set" .Values.operators.platformOperator.configuration.cluster.cloudflare.zoneId}}

- name: CLOUDFLARE_DOMAIN
  {{- if .Values.operators.platformOperator.configuration.cluster.cloudflare.baseDomain}}
  value: {{.Values.operators.platformOperator.configuration.cluster.cloudflare.baseDomain}}
  {{- else}}
    {{fail ".Values.operators.platformOperator.configuration.cluster.cloudflare.baseDomain must be set"}}
  {{- end}}

- name: KL_S3_IAC_BUCKET_NAME
  {{- if not $clusterOperatorVars.IACStateStore.s3BucketName}}}
  {{ fail ".operators.platformOperator.configuration.cluster.IACStateStore.s3BucketName is required" }}
  {{- end }}
  value: {{$clusterOperatorVars.IACStateStore.s3BucketName}}

- name: KL_S3_IAC_BUCKET_REGION
  value: {{ required ".Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketRegion must be set" .Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketRegion }}

- name: KL_S3_IAC_DIRECTORY
  value: {{ required ".Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketDir must be set" .Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketDir }} 

- name: MESSAGE_OFFICE_GRPC_ADDR
  value: "message-office.{{.Values.global.baseDomain}}:443"

- name: KL_AWS_ACCESS_KEY
  value: {{ required ".Values.aws.accessKey" .Values.aws.accessKey }} 

- name: KL_AWS_SECRET_KEY
  value: {{ required ".Values.aws.secretKey" .Values.aws.secretKey }} 

- name: IAC_JOB_IMAGE
  value: {{$clusterOperatorVars.jobImage }}

{{- end -}}
