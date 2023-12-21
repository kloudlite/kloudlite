---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: auth
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - auth.{{.Values.global.baseDomain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    {{if .Values.global.isDev}}
    - app: auth-web
      path: /ping
      port: 8000
      rewrite: false
    - app: auth-web
      path: /socket
      port: 8000
      rewrite: false
    {{end}}
    - app: auth-web
      path: /
      port: 80
---
