{{- if (eq .Values.cloudprovider "gcp") }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "gcp-credentials-secret-name" . }}
  namespace: {{.Release.Namespace}}
data:
  gcloud-creds.json: {{ .Values.gcp.gcloudServiceAccountCreds.json | squote }}
{{- end }}
