{{- $secret := (lookup "v1" "Secret" .Release.Namespace (include "apps.gatewayKubeReverseProxy.secret.name" .)) -}}

apiVersion: v1
kind: Secret
metadata:
  name: {{ include "apps.gatewayKubeReverseProxy.secret.name" . }}
  namespace: {{ .Release.Namespace }}
data:
  {{include "apps.gatewayKubeReverseProxy.secret.key" .}}: |+
    {{- if $secret }}
    {{ dig "data" (include "apps.gatewayKubeReverseProxy.secret.key" .) "." $secret }}
    {{- else }}
    {{ randBytes 64 | sha256sum | b64enc }}
    {{- end }}
