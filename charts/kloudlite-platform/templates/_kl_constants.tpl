{{- define "cert-manager.name" -}} cert-manager {{- end -}}
{{- define "cert-manager.chart.version" -}} v1.13.2 {{- end -}}

{{- define "nginx-ingress.name" -}} nginx-ingress {{- end -}}
{{- define "nginx-ingress.chart.version" -}} 4.9.0 {{- end -}}

{{- define "kloudlite.account-cookie-name" -}} kloudlite-account {{- end -}}
{{- define "kloudlite.cluster-cookie-name" -}} kloudlite-cluster {{- end -}}

{{- define "edge-gateways.secret.name" -}} kloudlite-edge-gateways {{- end -}}
{{- define "edge-gateways.secret.key" -}} gateways.yml {{- end -}}

{{- /* # apps */}}
{{- define "apps.accountsApi.name" -}} accounts-api {{- end -}}
{{- define "apps.accountsApi.httpPort" -}} 3000 {{- end -}}
{{- define "apps.accountsApi.grpcPort" -}} 3001 {{- end -}}

{{- define "apps.iamApi.name" -}} iam {{- end -}}
{{- define "apps.iamApi.grpcPort" -}} 3001 {{- end -}}

{{- define "apps.commsApi.name" -}} comms-api {{- end -}}
{{- define "apps.commsApi.httpPort" -}} 3000 {{- end -}}
{{- define "apps.commsApi.grpcPort" -}} 3001 {{- end -}}

{{- define "apps.consoleApi.name" -}} console-api {{- end -}}
{{- define "apps.consoleApi.httpPort" -}} 3000 {{- end -}}
{{- define "apps.consoleApi.grpcPort" -}} 3001 {{- end -}}
{{- define "apps.consoleApi.dnsPort" -}} 5353 {{- end -}}

{{- define "apps.infraApi.name" -}} infra-api {{- end -}}
{{- define "apps.infraApi.httpPort" -}} 3000 {{- end -}}
{{- define "apps.infraApi.grpcPort" -}} 3001 {{- end -}}

{{- define "apps.authApi.name" -}} auth-api {{- end -}}
{{- define "apps.authApi.httpPort" -}} 3000 {{- end -}}
{{- define "apps.authApi.grpcPort" -}} 3001 {{- end -}}
{{- define "apps.authApi.oAuth2-secret.name" -}} oauth2-secrets {{- end -}}

{{- define "apps.messageOffice.name" -}} message-office {{- end -}}
{{- define "apps.messageOffice.httpPort" -}} 3000{{- end -}}
{{- define "apps.messageOffice.privateGrpcPort" -}} 3002 {{- end -}}
{{- define "apps.messageOffice.publicGrpcPort" -}} 3001 {{- end -}}

{{- define "apps.messageOffice.token-hasing.secret.name" -}} messaage-office-token-hashing {{- end -}}
{{- define "apps.messageOffice.token-hasing.secret.key" -}} token {{- end -}}

{{- define "apps.gatewayApi.name" -}} gateway-api {{- end -}}
{{- define "apps.gatewayApi.httpPort" -}} 3000 {{- end -}}

{{- define "apps.observabilityApi.name" -}} observability-api {{- end -}}
{{- define "apps.observabilityApi.httpPort" -}} 3000 {{- end -}}

{{- define "apps.webhooksApi.name" -}} webhooks-api {{- end -}}
{{- define "apps.webhooksApi.httpPort" -}} 3000 {{- end -}}
{{- define "apps.webhooksApi.authenticationSecret.name" -}} webhook-authn-secrets {{- end -}}
{{- define "apps.webhooksApi.authenticationSecret.token-key" -}} webhook-authn-secrets {{- end -}}

{{- define "apps.websocketApi.name" -}} websocket-api {{- end -}}
{{- define "apps.websocketApi.httpPort" -}} 3000 {{- end -}}

{{- define "apps.klInstaller.name" -}} kl-installer {{- end -}}
{{- define "apps.klInstaller.httpPort" -}} 3000 {{- end -}}

{{- define "apps.authWeb.name" -}} auth-web {{- end -}}
{{- define "apps.authWeb.httpPort" -}} 3000 {{- end -}}

{{- define "apps.consoleWeb.name" -}} console-web {{- end -}}
{{- define "apps.consoleWeb.httpPort" -}} 3000 {{- end -}}

{{- define "apps.healthApi.name" -}} health-api {{- end -}}
{{- define "apps.healthApi.httpPort" -}} 3000 {{- end -}}

{{- define "apps.gatewayKubeReverseProxy.secret.name" -}} gvpn-gateway-reverse-proxy-authz {{- end -}}
{{- define "apps.gatewayKubeReverseProxy.secret.key" -}} authz-token {{- end -}}

{{- define "self-edge-gateway.public.host" -}} wg-gateways.{{.Values.baseDomain}} {{- end -}}

{{- /* mongodb databases */}}

{{- define "mongo.auth-db" -}} auth-db {{- end -}}
{{- define "mongo.console-db" -}} console-db {{- end -}}
{{- define "mongo.comms-db" -}} comms-db {{- end -}}
{{- define "mongo.accounts-db" -}} accounts-db {{- end -}}
{{- define "mongo.registry-db" -}} registry-db {{- end -}}
{{- define "mongo.events-db" -}} events-db {{- end -}}
{{- define "mongo.iam-db" -}} iam-db {{- end -}}
{{- define "mongo.infra-db" -}} infra-db {{- end -}}
{{- define "mongo.message-office-db" -}} message-office-db {{- end -}}

{{- /* helm charts */}}
{{- define "nats.name" -}} nats {{- end -}}
{{- define "nats.url" -}} nats://{{include "nats.name" .}}:4222 {{- end -}}

{{- define "vector.name" -}} vector {{- end -}}
{{- define "vector.grpc-addr" -}} {{- include "vector.name" .}}:6000 {{- end -}}
{{- define "vector.chart.version" -}} 0.23.0 {{- end -}}

{{- define "vector-agent.name" -}} vector {{- end -}}
{{- define "vector-agent.chart.verison" -}} 0.30.0 {{- end -}}

{{- define "victoria-metrics.name" -}} victoria-metrics {{- end -}}
{{- define "victoria-metrics.chart.version" -}} 0.18.11 {{- end -}}
{{- define "victoria-metrics.prom-url" -}}
http://vmselect-{{ include "victoria-metrics.name" .}}.{{$.Release.Namespace}}.svc.{{$.Values.clusterInternalDNS}}:8481/select/0/prometheus
{{- end -}}
