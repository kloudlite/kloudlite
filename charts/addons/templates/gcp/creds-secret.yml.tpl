{{- if (and (eq .Values.cloudprovider "gcp") .Values.gcp.csi_driver.enabled) }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "gcp-credentials-secret-name" . }}
  namespace: {{.Release.Namespace}}
data:
  gcloud-creds.json: {{ .Values.gcp.gcloudServiceAccountCreds.json | squote }}
{{- end }}
