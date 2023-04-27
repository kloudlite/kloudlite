apiVersion: wg.kloudlite.io/v1
kind: Region
metadata:
  name: {{.Values.accountName}}-{{.Values.defaultRegion}}
spec:
  accountName: {{.Values.accountName}}
  isMaster: true
