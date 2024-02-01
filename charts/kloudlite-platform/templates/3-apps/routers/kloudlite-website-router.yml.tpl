---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: kloudlite-website
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - {{include "router-domain" .}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: kloudlite-website
      path: /
      port: 80
---
