---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: console
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - "console.{{.Values.global.baseDomain}}"
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
        port: 80
        rewrite: false
---
