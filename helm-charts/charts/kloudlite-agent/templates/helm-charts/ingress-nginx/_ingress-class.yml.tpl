apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: {{.Values.helmCharts.ingressNginx.configuration.ingressClassName}}
spec:
  controller: k8s.io/{{.Values.helmCharts.ingressNginx.configuration.ingressClassName}}

