{{- $name := .Release.Name -}}
{{- $namespace := .Release.Namespace -}}

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    {{if .Values.distribution.tls.enabled }}
    cert-manager.io/cluster-issuer: {{.Values.certManager.certIssuer.name}}
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.kubernetes.io/ssl-redirect: "true"
    {{end}}
    nginx.ingress.kubernetes.io/preserve-trailing-slash: "true"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
    nginx.ingress.kubernetes.io/rewrite-target: /$1
    nginx.ingress.kubernetes.io/proxy-body-size: "0" 
    nginx.ingress.kubernetes.io/proxy-buffer-size: 100m
    nginx.ingress.kubernetes.io/proxy-buffers-number: "8"
    nginx.ingress.kubernetes.io/proxy-busy-buffers-size: 100m
    nginx.ingress.kubernetes.io/proxy-busy-buffers-number: "8"
    nginx.ingress.kubernetes.io/proxy-max-temp-file-size: "0" 
    nginx.ingress.kubernetes.io/auth-url: http://container-registry-api.kl-core.svc.cluster.local:4000/auth?path=$request_uri&method=$request_method

  name: {{ $name }}
spec:
  {{ if .Values.distribution.tls.enabled }}
  tls:
    - hosts:
      - "*.{{ .Values.global.baseDomain }}"
      - "*.{{ .Values.global.secondaryDomain }}"
  {{ end }}
  ingressClassName: {{ .Values.global.ingressClassName }}
  rules:
  - host: registry.{{ .Values.global.baseDomain }}
    http:
      paths:
      - backend:
          service:
            name: svc-{{ $name }}
            port:
              number: 80
        path: /(.*)
        pathType: Prefix
