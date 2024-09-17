---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: observe
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.nginxIngress.ingressClass.name }}
  domains:
    - observe.{{.Values.baseDomain}}
  cors:
    enabled: true
    origins:
      - https://.Values.baseDomain}}
      - https://console.{{.Values.baseDomain}}
    allowCredentials: true
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: observability-api
      path: /
      port: {{ include "apps.observabilityApi.httpPort" . }}
---
