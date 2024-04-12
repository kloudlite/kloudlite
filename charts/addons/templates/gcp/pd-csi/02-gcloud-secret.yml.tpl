{{- if (and (eq .Values.cloudprovider "gcp") .Values.gcp.csi_driver.enabled) }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "gcp-credentials-secret-name" . }}
  namespace: {{include "gcp-csi-namespace" .}}
data:
  cloud-sa.json: {{ .Values.gcp.gcloudServiceAccountCreds.json | squote }}
{{- end }}
