---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: message-office
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  backendProtocol: GRPC
  maxBodySizeInMB: 50
  domains:
    - message-office.{{include "router-domain" .}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: message-office
      path: /
      port: {{.Values.apps.messageOfficeApi.configuration.externalGrpcPort}}
---
