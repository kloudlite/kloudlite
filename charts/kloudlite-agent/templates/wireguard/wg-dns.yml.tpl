apiVersion: wireguard.kloudlite.io/v1
kind: Dns
metadata:
  name: {{.Release.Name}}-wg-dns
  namespace: {{.Release.Namespace}}
spec: {}
