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
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: websocket-api
      path: /ws
      port: 80
      rewrite: false
    - app: auth-web
      path: /logs
      port: 80
      rewrite: false
---
