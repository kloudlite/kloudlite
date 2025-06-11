---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: scripts
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.nginxIngress.ingressClass.name }}
  domains:
    - scripts.{{.Values.baseDomain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: webhooks-api
      path: /image-hook.sh
      port: {{ include "apps.webhooksApi.httpPort" . }}
---
