{{- define "networking-operator-env" -}}
- name: IMAGE_WEBHOOK_SERVER
  value: {{.Values.agentOperator.configuration.gateway.imageWebhookServer.repository}}:{{.Values.agentOperator.configuration.gateway.imageWebhookServer.tag | default (include "image-tag" .) }}

- name: IMAGE_IP_MANAGER
  value: {{.Values.agentOperator.configuration.gateway.imageIPManager.repository}}:{{.Values.agentOperator.configuration.gateway.imageIPManager.tag | default (include "image-tag" .) }}

- name: IMAGE_IP_BINDING_CONTROLLER
  value: {{.Values.agentOperator.configuration.gateway.imageIPBindingController.repository}}:{{.Values.agentOperator.configuration.gateway.imageIPBindingController.tag | default (include "image-tag" .) }}

- name: IMAGE_DNS
  value: {{.Values.agentOperator.configuration.gateway.imageDNS.repository}}:{{.Values.agentOperator.configuration.gateway.imageDNS.tag | default (include "image-tag" .) }}

- name: IMAGE_LOGS_PROXY
  value: {{.Values.agentOperator.configuration.gateway.imageLogsProxy.repository}}:{{.Values.agentOperator.configuration.gateway.imageLogsProxy.tag | default (include "image-tag" .) }}
{{- end -}}

