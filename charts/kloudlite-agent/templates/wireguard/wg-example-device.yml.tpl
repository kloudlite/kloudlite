apiVersion: wireguard.kloudlite.io/v1
kind: Device
metadata:
  labels:
    kloudlite.io/account.name: {{.Values.accountName}}
    kloudlite.io/cluster.name: {{.Values.clusterName}}
  name: example-device
spec:
  accountName: {{.Values.accountName}}
  clusterName: {{.Values.clusterName}}
  ports:
  - port: 80
    targetPort: 3000
