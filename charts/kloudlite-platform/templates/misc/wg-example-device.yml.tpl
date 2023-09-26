apiVersion: wireguard.kloudlite.io/v1
kind: Device
metadata:
  name: example-device
spec:
  offset: 1
  ports:
  - port: 80
    targetPort: 3000
  - port: 9100
    targetPort: 9999
  - port: 3001
    targetPort: 3001
  serverName: platform
