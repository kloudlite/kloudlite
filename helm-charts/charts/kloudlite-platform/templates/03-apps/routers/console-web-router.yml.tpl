---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: console
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.nginxIngress.ingressClass.name }}
  domains:
    - console.{{.Values.baseDomain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: console-web
      path: /
      port: {{ include "apps.consoleWeb.httpPort" . }}
      rewrite: false
---
