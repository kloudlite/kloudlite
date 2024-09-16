apiVersion: v1
kind: Secret
metadata:
  name: {{ include "edge-gateways.secret.name" . }}
  namespace: {{ .Release.Namespace }}
stringData:
  gateways.yml: |+
    {{- .Values.edgeGateways | toYaml | nindent 4 }}
