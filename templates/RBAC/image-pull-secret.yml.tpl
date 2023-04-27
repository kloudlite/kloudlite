apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.rbac.pullSecret.name}}
  namespace: {{.Release.Namespace}}
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: {{.Values.rbac.pullSecret.value}}
