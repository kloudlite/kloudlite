{{- if .Values.operators.platformOperator.configuration.wireguard.enableExamples }}
apiVersion: v1
kind: Namespace
metadata:
  name: kl-vpn-devices
---
apiVersion: wireguard.kloudlite.io/v1
kind: Device
metadata:
  name: example-device
  namespace: kl-vpn-devices
spec:
  ports:
  - port: 80
    targetPort: 3000
  - port: 3001
    targetPort: 3001
{{- end }}
