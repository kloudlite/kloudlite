{{- if .Values.apps.gatewayApi.exposeWithIngress }}

---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: gateway-api
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.nginxIngress.ingressClass.name }}
  domains:
    - gateway-api.{{.Values.baseDomain}}
  https:
    enabled: true
    forceRedirect: true
  cors:
    enabled: true
    origins:
      - https://studio.apollographql.com
    allowCredentials: true
  basicAuth:
    enabled: true
    username: admin
  routes:
    - app: gateway
      path: /
      port: {{ include "apps.gatewayApi.httpPort" . }}

{{- end }}
