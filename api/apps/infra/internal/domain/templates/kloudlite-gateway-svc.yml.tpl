apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace |squote}}
  labels: 
    kloudlite.io/description: "kl-gateway-namespace"
---
apiVersion: v1
kind: Service
metadata:
  name: "{{.Name}}"
  namespace: "{{.Namespace}}"
  labels: 
    app: {{.Name}}
spec:
  type: NodePort
  ports:
    - name: wireguard
      protocol: UDP
      port: {{.WireguardPort}}
      targetPort: {{.WireguardPort}}
  selector: {{.Selector | toYAML | nindent 4 }}
