{{- define "networking-operator-env" -}}
- name: IMAGE_WEBHOOK_SERVER
  value: {{.Values.operators.platformOperator.configuration.gateway.imageWebhookServer.repository}}:{{.Values.operators.platformOperator.configuration.gateway.imageWebhookServer.tag | default (include "image-tag" .) }}

- name: IMAGE_IP_MANAGER
  value: {{.Values.operators.platformOperator.configuration.gateway.imageIPManager.repository}}:{{.Values.operators.platformOperator.configuration.gateway.imageIPManager.tag | default (include "image-tag" .) }}

- name: IMAGE_IP_BINDING_CONTROLLER
  value: {{.Values.operators.platformOperator.configuration.gateway.imageIPBindingController.repository}}:{{.Values.operators.platformOperator.configuration.gateway.imageIPBindingController.tag | default (include "image-tag" .) }}

- name: IMAGE_DNS
  value: {{.Values.operators.platformOperator.configuration.gateway.imageDNS.repository}}:{{.Values.operators.platformOperator.configuration.gateway.imageDNS.tag | default (include "image-tag" .) }}

- name: IMAGE_LOGS_PROXY
  value: {{.Values.operators.platformOperator.configuration.gateway.imageLogsProxy.repository}}:{{.Values.operators.platformOperator.configuration.gateway.imageLogsProxy.tag | default (include "image-tag" .) }}
{{- end -}}
