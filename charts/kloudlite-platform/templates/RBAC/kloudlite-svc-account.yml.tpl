apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Values.normalSvcAccount}}
  namespace: {{.Release.Namespace}}
imagePullSecrets:
  - name: {{.Values.rbac.pullSecret.name}}
