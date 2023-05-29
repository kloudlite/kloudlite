apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Values.normalSvcAccount}}
  namespace: {{.Release.Namespace}}
