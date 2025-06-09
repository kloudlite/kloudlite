{{- $domain := get . "domain" -}}
{{- $namespace := get . "namespace" -}}
{{- $ingressClass := get . "ingress-class" -}}

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: s1-edge-ingress
  namespace: {{$namespace}}
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "false"
spec:
  ingressClassName: ingress-nginx-acc-xnnbf-kd254lyrcnu2jqvnoksvk
  rules:
    - host: "{{$domain}}"
      http:
        paths:
          - backend:
              service:
                name: s1
                port:
                  number: 80
            path: /
            pathType: Prefix
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: s1
  name: s1
  namespace: {{$namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: s1
  template:
    metadata:
      labels:
        app: s1
    spec:
      nodeSelector:
        kloudlite.io/region: reg-9tco8kqki8hsw8dcwftdtkesxhqa
      tolerations:
        - effect: NoExecute
          key: kloudlite.io/region
          operator: Equal
          value: reg-9tco8kqki8hsw8dcwftdtkesxhqa
      dnsPolicy: ClusterFirst
      containers:
      - image: kennethreitz/httpbin:latest
#       - image: k8s.gcr.io/e2e-test-images/jessie-dnsutils:1.3
        imagePullPolicy: IfNotPresent
        name: main
#         command:
#           - sh
#           - -c
#           - tail -f /dev/null
---
apiVersion: v1
kind: Service
metadata:
  name: s1
  namespace: {{$namespace}}
spec:
  ports:
  - port: 80
    protocol: TCP
  selector:
    app: s1
---
