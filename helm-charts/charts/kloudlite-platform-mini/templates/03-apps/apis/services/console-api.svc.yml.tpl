apiVersion: v1
kind: Service
metadata:
  name: {{ include "apps.consoleApi.name" . }}-dns
  namespace: {{.Release.Namespace}}
spec:
  type: LoadBalancer
  selector:
    kloudlite.io/app.name: {{ include "apps.consoleApi.name" . }}
  ports:
  - port: 53
    targetPort: {{ include "apps.consoleApi.dnsPort" . }}
    protocol: UDP
    name: dns
