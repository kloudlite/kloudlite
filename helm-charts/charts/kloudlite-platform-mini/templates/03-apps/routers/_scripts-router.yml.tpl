---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: scripts
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.ingress.ingressClass }}
  domains:
    - scripts.{{.Values.webHost}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: webhooks-api
      path: /image-hook.sh
      port: {{ include "apps.webhooksApi.httpPort" . }}
---
