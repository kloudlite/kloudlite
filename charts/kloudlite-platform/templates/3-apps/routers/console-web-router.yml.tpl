---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: console
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - console.{{include "router-domain" .}}
  https:
    enabled: true
    forceRedirect: true
  routes:
       {{if .Values.global.isDev}}
      - app: console-web
        path: /ping
        port: 8001
        rewrite: false
      - app: console-web
        path: /socket
        port: 8001
        rewrite: false
      {{end}}
      - app: console-web
        path: /
        port: {{.Values.apps.consoleWeb.configuration.httpPort}}
        rewrite: false
---
