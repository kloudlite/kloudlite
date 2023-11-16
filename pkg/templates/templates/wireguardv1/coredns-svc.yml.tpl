{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}

apiVersion: v1
kind: Service
metadata:
  name: {{ $name }}
  namespace: {{ $namespace }}
spec:
  selector:
    app: dns
  ports:
    - name: dns
      protocol: UDP
      port: 53
    - name: dns-tcp
      protocol: TCP
      port: 53
