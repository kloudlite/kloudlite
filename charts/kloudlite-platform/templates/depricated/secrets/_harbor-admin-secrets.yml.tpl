{{- if .Values.operators.artifactsHarbor.enabled }}

---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secretNames.harborAdminSecret}}
  namespace: {{.Release.Namespace}}
stringData:
  ADMIN_USERNAME: {{.Values.operators.artifactsHarbor.configuration.adminUsername}}
  ADMIN_PASSWORD: {{.Values.operators.artifactsHarbor.configuration.adminPassword}}
  IMAGE_REGISTRY_HOST: {{.Values.operators.artifactsHarbor.configuration.imageRegistryHost}}
---
{{- end }}
