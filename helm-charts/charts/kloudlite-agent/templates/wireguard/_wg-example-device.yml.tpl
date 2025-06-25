apiVersion: wireguard.kloudlite.io/v1
kind: Device
metadata:
  name: example-device
spec:
  ports:
  - port: 80
    targetPort: 3000
