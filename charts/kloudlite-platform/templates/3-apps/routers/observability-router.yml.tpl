---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: observe
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - observe.{{include "router-domain" .}}
  cors:
    enabled: true
    origins:
      - https://{{include "router-domain" .}}
      - https://console.{{include "router-domain" .}}
    allowCredentials: true
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: observability-api
      path: /
      port: {{.Values.apps.observabilityApi.configuration.httpPort}}
---
