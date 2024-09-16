apiVersion: v1
kind: Secret
metadata:
  name: {{ include "edge-gateways.secret.name" . }}
  namespace: {{ .Release.Namespace }}
stringData:
  gateways.yml: |+
    {{- range $k, $v := .Values.edgeGateways }}
    - {{ merge $v (eq $v.id "self" | ternary (dict "publicDNSHost" (include "self-edge-gateway.public.host" $)) dict ) | toJson }}
    {{- end }}
