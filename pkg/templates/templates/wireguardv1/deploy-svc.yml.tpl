{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}

kind: Service
apiVersion: v1
metadata:
  annotations:
    kloudlite.io/device.name: {{$name}}
  labels:
    k8s-app: wireguard
    kloudlite.io/wg-service: "true"
    kloudlite.io/wg-device.name: {{ $name }}
  name: wg-server-{{$name}}
  namespace: {{$namespace}}
spec:
  type: NodePort
  ports:
    - port: 51820
      protocol: UDP
      targetPort: 51820
  selector:
    kloudlite.io/pod-type: wireguard-server
    kloudlite.io/device: {{$name}}
