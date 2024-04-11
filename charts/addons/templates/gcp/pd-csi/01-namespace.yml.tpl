{{- if (and (eq .Values.cloudprovider "gcp") .Values.gcp.csi_driver.enabled) }}
apiVersion: v1
kind: Namespace
metadata:
  name: {{ include "gcp-csi-namespace" . | squote}}
{{- end}}
