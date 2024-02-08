---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: websocket-api
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - websocket.{{include "router-domain" .}}
  cors:
    enabled: true
    origins:
      - https://kloudlite.io
      - https://console.{{include "router-domain" . }}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: websocket-api
      path: /ws
      port: 80
      rewrite: false
    - app: websocket-api
      path: /logs
      port: 80
      rewrite: false
---
