---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: websocket-api
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.ingress.ingressClass }}
  domains:
    - websocket.{{.Values.webHost}}
  cors:
    enabled: true
    origins:
      - https://kloudlite.io
      - https://console.{{.Values.webHost }}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: websocket-api
      path: /ws
      port: {{ include "apps.websocketApi.httpPort" . }}
      rewrite: false
    - app: websocket-api
      path: /logs
      port: {{ include "apps.websocketApi.httpPort" . }}
      rewrite: false
---
