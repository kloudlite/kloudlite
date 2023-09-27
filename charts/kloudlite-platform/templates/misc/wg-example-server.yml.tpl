{{- if .Values.operators.wgOperator.configuration.enableExamples }}
apiVersion: wireguard.kloudlite.io/v1
kind: Server
metadata:
  name: platform
spec:
  accountName: kloudlite
  clusterName: platform
{{- end }}
