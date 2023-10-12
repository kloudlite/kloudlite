{{- if .Values.operators.wgOperator.configuration.enableExamples }}
apiVersion: wireguard.kloudlite.io/v1
kind: Server
metadata:
  name: platform
spec:
  accountName: kloudlite
  clusterName: platform
  publicKey: /4DhvDgf5yo7dJG3ChB+Oy7y7s8T/0gj08eaUD0/3R0=
---
apiVersion: v1
kind: Namespace
metadata:
  name: wg-platform
{{- end }}
