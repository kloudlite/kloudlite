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
  value: {{.Values.envVars.nats.url}}

- name: NATS_STREAM
  value: {{.Values.envVars.nats.streams.resourceSync.name}}

- name: NATS_CLUSTER_UPDATE_SUBJECT_FORMAT
  value: "{{.Values.envVars.nats.streams.resourceSync.name}}.account-%s.cluster-%s.platform.kloudlite-infra.resource-update"

- name: CLOUDFLARE_ZONE_ID
  value: {{ required ".Values.operators.platformOperator.configuration.cluster.cloudflare.zoneId must be set" .Values.operators.platformOperator.configuration.cluster.cloudflare.zoneId}}

- name: CLOUDFLARE_DOMAIN
  {{- if .Values.operators.platformOperator.configuration.cluster.cloudflare.baseDomain}}
  value: {{.Values.operators.platformOperator.configuration.cluster.cloudflare.baseDomain}}
  {{- else}}
    {{fail ".Values.operators.platformOperator.configuration.cluster.cloudflare.baseDomain must be set"}}
  {{- end}}

- name: MESSAGE_OFFICE_GRPC_ADDR
  value: message-office.{{include "router-domain" .}}:443

- name: KL_AWS_ACCESS_KEY
  value: {{ required ".Values.aws.accessKey" .Values.aws.accessKey }}

- name: KL_AWS_SECRET_KEY
  value: {{ required ".Values.aws.secretKey" .Values.aws.secretKey }}

- name: IAC_JOB_IMAGE
  value: {{.Values.operators.platformOperator.configuration.cluster.jobImage.repository}}:{{.Values.operators.platformOperator.configuration.cluster.jobImage.tag | default (include "image-tag" .) }}

- name: "IAC_JOB_TOLERATIONS"
  value: {{.Values.nodepools.iac.tolerations | toJson | squote}}

- name: IAC_JOB_NODE_SELECTOR
  value: {{.Values.nodepools.iac.labels | toJson |squote}}

{{- end -}}
