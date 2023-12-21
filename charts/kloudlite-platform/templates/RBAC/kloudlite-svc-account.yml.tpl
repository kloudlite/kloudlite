apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Values.global.normalSvcAccount}}
  namespace: {{.Release.Namespace}}
