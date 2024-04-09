---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: iot-console-api
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - iotnet.{{include "router-domain" .}}
  cors:
    enabled: true
    origins:
      - https://kloudlite.io
      - https://iotnet.{{include "router-domain" . }}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: iot-console-api
      path: /
      port: 80
      rewrite: false
---

