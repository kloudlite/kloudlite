apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.clusterIdentitySecretName}}
  namespace: {{.Release.Namespace}}
stringData:
  CLUSTER_TOKEN: ""
  ACCESS_TOKEN: ""
