---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: message-office
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.ingress.ingressClass }}
  backendProtocol: GRPC
  maxBodySizeInMB: 50
  domains:
    - message-office.{{.Values.webHost}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: message-office
      path: /
      port: {{ include "apps.messageOffice.publicGrpcPort" . }}
---
