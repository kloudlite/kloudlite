---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: websocket-api
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - websocket.{{.Values.global.baseDomain}}
  cors:
    enabled: true
    origins:
      - https://kloudlite.io
      - https://console.{{.Values.global.baseDomain}}
      {{- /* - https://studio.apollographql.com */}}
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
