---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.operators.platformOperator.name}}-cloudflare-params
  namespace: {{.Release.Namespace}}
data:
  api_token: {{.Values.operators.platformOperator.configuration.cluster.cloudflare.apiToken | b64enc | quote }}
  base_domain: {{.Values.operators.platformOperator.configuration.cluster.cloudflare.baseDomain | b64enc | quote }}
  zone_id: {{.Values.operators.platformOperator.configuration.cluster.cloudflare.zoneId | b64enc | quote }}
---

{{- define "cluster-operator-env" -}}
{{- $operatorName := .Values.operators.platformOperator.name }} 
{{- $clusterOperatorVars := .Values.operators.platformOperator.configuration.cluster }} 

- name: CLOUDFLARE_API_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{$operatorName}}-cloudflare-params
      key: api_token

- name: CLOUDFLARE_ZONE_ID
  valueFrom:
    secretKeyRef:
      name: {{$operatorName}}-cloudflare-params
      key: zone_id

- name: CLOUDFLARE_DOMAIN
  valueFrom:
    secretKeyRef:
      name: {{$operatorName}}-cloudflare-params
      key: base_domain

- name: KL_S3_BUCKET_NAME
  {{- if not $clusterOperatorVars.IACStateStore.s3BucketName}}}
  {{ fail ".operators.platformOperator.configuration.cluster.IACStateStore.s3BucketName is required" }}
  {{- end }}
  value: {{$clusterOperatorVars.IACStateStore.s3BucketName}}

- name: KL_S3_BUCKET_REGION
  value: {{ required ".Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketRegion must be set" .Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketRegion }}

- name: KL_S3_BUCKET_DIRECTORY
  value: {{ required ".Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketDir must be set" .Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketDir }} 

- name: MESSAGE_OFFICE_GRPC_ADDR
  value: "{{.Values.routers.messageOfficeApi.name}}.{{.Values.baseDomain}}:443"

- name: KL_AWS_ACCESS_KEY
  value: {{ required ".Values.operators.platformOperator.configuration.cluster.IACStateStore.accessKey must be set" .Values.operators.platformOperator.configuration.cluster.IACStateStore.accessKey }} 

- name: KL_AWS_SECRET_KEY
  value: {{ required ".Values.operators.platformOperator.configuration.cluster.IACStateStore.secretKey must be set" .Values.operators.platformOperator.configuration.cluster.IACStateStore.secretKey }} 

- name: IAC_JOB_IMAGE
  value: {{$clusterOperatorVars.jobImage.name}}:{{$clusterOperatorVars.jobImage.tag | default .Values.kloudlite_release }}

{{- end -}}
