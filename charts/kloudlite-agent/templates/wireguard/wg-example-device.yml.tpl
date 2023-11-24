apiVersion: wireguard.kloudlite.io/v1
kind: Device
metadata:
  labels:
    kloudlite.io/account.name: {{.Values.accountName}}
    kloudlite.io/cluster.name: {{.Values.clusterName}}
    kloudlite.io/wg-server.name: {{.Release.Name}}-wg-dns
  name: example-device
  namespace: {{.Release.Namespace}}
spec:
  accountName: {{.Values.accountName}}
  clusterName: {{.Values.clusterName}}
  ports:
  - port: 80
    targetPort: 3000
