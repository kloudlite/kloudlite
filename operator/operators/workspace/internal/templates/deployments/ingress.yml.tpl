---
{{- with . }}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: code-server-{{.Metadata.Name}}
  namespace: {{.Metadata.Namespace}}
  labels: {{.Metadata.Labels | toJson }}
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
spec:
  {{- if and .IngressClassName (ne .IngressClassName "") }}
  ingressClassName: {{.IngressClassName}}
  {{- end }}
  rules:
  - host: code-server.{{.Metadata.Name}}.{{.WorkMachineName}}.{{.KloudliteDomain}}
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: {{.Metadata.Name}}
            port:
              number: {{.PortConfig.CodeServerPort}}


---


apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nb-{{.Metadata.Name}}
  namespace: {{.Metadata.Namespace}}
  labels: {{.Metadata.Labels | toJson }}
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
spec:
  {{- if and .IngressClassName (ne .IngressClassName "") }}
  ingressClassName: {{.IngressClassName}}
  {{- end }}
  rules:
  - host: notebook.{{.Metadata.Name}}.{{.WorkMachineName}}.{{.KloudliteDomain}}
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: {{.Metadata.Name}}
            port:
              number: {{.PortConfig.NotebookPort}}


---

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ttyd-{{.Metadata.Name}}
  namespace: {{.Metadata.Namespace}}
  labels: {{.Metadata.Labels | toJson }}
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
spec:
  {{- if and .IngressClassName (ne .IngressClassName "") }}
  ingressClassName: {{.IngressClassName}}
  {{- end }}
  rules:
  - host: ttyd.{{.Metadata.Name}}.{{.WorkMachineName}}.{{.KloudliteDomain}}
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: {{.Metadata.Name}}
            port:
              number: {{.PortConfig.TTYDPort}}

{{- end }}
