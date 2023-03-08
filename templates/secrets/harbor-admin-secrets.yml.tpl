---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secrets.names.harborAdminSecret}}
  namespace: {{.Release.Namespace}}
stringData:
  ADMIN_USERNAME: {{.Values.harbor.adminUsername}}
  ADMIN_PASSWORD: {{.Values.harbor.adminPassword}}
  IMAGE_REGISTRY_HOST: {{.Values.harbor.imageRegistryHost}}
---
