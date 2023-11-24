---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.operators.clusterOperator.name}}-cloudflare-params
  namespace: {{.Release.Namespace}}
data:
  api_token: {{.Values.operators.clusterOperator.configuration.cloudflare.apiToken | b64enc | quote }}
  base_domain: {{.Values.operators.clusterOperator.configuration.cloudflare.baseDomain | b64enc | quote }}
  zone_id: {{.Values.operators.clusterOperator.configuration.cloudflare.zoneId | b64enc | quote }}
---

{{- define "cluster-operator-env" -}}
{{- $clusterOperator := .Values.operators.clusterOperator }} 

- name: CLOUDFLARE_API_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{$clusterOperator.name}}-cloudflare-params
      key: api_token

- name: CLOUDFLARE_ZONE_ID
  valueFrom:
    secretKeyRef:
      name: {{$clusterOperator.name}}-cloudflare-params
      key: zone_id

- name: CLOUDFLARE_DOMAIN
  valueFrom:
    secretKeyRef:
      name: {{$clusterOperator.name}}-cloudflare-params
      key: base_domain

- name: KL_S3_BUCKET_NAME
  {{- if not $clusterOperator.configuration.IACStateStore.s3BucketName}}}
  {{ fail ".operators.clusterOperator.configuration.IACStateStore.s3BucketName is required" }}
  {{- end }}
  value: {{$clusterOperator.configuration.IACStateStore.s3BucketName}}

- name: KL_S3_BUCKET_REGION
  {{- if not $clusterOperator.configuration.IACStateStore.s3BucketRegion}}}
  {{ fail ".operators.clusterOperator.configuration.IACStateStore.s3BucketRegion is required" }}
  {{- end }}
  value: {{.Values.operators.clusterOperator.configuration.IACStateStore.s3BucketRegion}}

- name: MESSAGE_OFFICE_GRPC_ADDR
  value: "{{.Values.routers.messageOfficeApi.name}}.{{.Values.baseDomain}}:443"

- name: KL_AWS_ACCESS_KEY
{{- if not $clusterOperator.configuration.IACStateStore.accessKey}}}
{{ fail ".operators.clusterOperator.configuration.IACStateStore.accessKey is required" }}
{{- end }}
  value: "{{$clusterOperator.configuration.IACStateStore.accessKey}}"

- name: KL_AWS_SECRET_KEY
{{- if not $clusterOperator.configuration.IACStateStore.secretKey}}}
{{ fail ".operators.clusterOperator.configuration.IACStateStore.secretKey is required" }}
{{- end }}
  value: "{{ $clusterOperator.configuration.IACStateStore.secretKey }}"

- name: IAC_JOB_IMAGE
  value: {{.Values.operators.clusterOperator.configuration.jobImage.name}}:{{.Values.operators.clusterOperator.configuration.jobImage.tag | default .Values.kloudlite_release }}

{{- end -}}
