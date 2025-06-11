---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: message-office
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.nginxIngress.ingressClass.name }}
  backendProtocol: GRPC
  maxBodySizeInMB: 50
  domains:
    - message-office.{{.Values.baseDomain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: message-office
      path: /
      port: {{ include "apps.messageOffice.publicGrpcPort" . }}
---
