---
apiVersion: v1
kind: Secret
metadata:
  name: kloudlite-cloudflare
  namespace: {{.Release.Namespace}}
stringData:
    api_token: {{.Values.cloudflareWildCardCert.cloudflareCreds.secretToken}}

---

{{- if .Values.operators.platformOperator.configuration.nodepool.extractFromCluster }}
{{- $k3sParams := (lookup "v1" "Secret" "kube-system" "k3s-params") -}}

{{- if not $k3sParams }}
{{ fail "secret k3s-params not found in namespace kube-system, could not proceed with chart installation" }}
{{- end }}

apiVersion: v1
kind: Secret
metadata:
  name: k3s-params
  namespace: {{.Release.Namespace}}
data: {{ $k3sParams.data | toYaml | nindent 2 }}
{{- end }}
