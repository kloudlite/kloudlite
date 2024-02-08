---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: kl-installer
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - kl.{{include "router-domain" .}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: kl-installer
      path: /
      port: 80
      rewrite: false
---
