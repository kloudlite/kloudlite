---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: observe
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.ingress.ingressClass }}
  domains:
    - observe.{{.Values.webHost}}
  cors:
    enabled: true
    origins:
      - https://.Values.webHost}}
      - https://console.{{.Values.webHost}}
    allowCredentials: true
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: observability-api
      path: /
      port: {{ include "apps.observabilityApi.httpPort" . }}
---
