{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}
{{- $ownerRefs := get . "ownerRefs"}}

apiVersion: v1
kind: Service
metadata:
  name: kl-coredns
  namespace: {{ $namespace }}
  ownerReferences: {{ $ownerRefs| toJson}}
  labels:
    kloudlite.io/coredns-svc: {{ $name }}
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
