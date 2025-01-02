{{- $domains := get . "domains" }}

{{- $clusterIssuer := get . "cluster-issuer" }}
{{- $wildcardDomainSuffix := get . "wildcard-domain-suffix"}}

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{.Namespace}}-ingress
  namespace: {{.Namespace}}
  annotations:
    {{ K8sAnnotation true "cert-manager.io/cluster-issuer" $clusterIssuer }}

    {{- $bodySize := printf "%dm" .Spec.MaxBodySizeInMB}}
    {{K8sAnnotation .Spec.MaxBodySizeInMB "nginx.ingress.kubernetes.io/proxy-body-size" $bodySize }}
    {{K8sAnnotation true "nginx.ingress.kubernetes.io/ssl-redirect" .Spec.Https.Enabled }}
    {{K8sAnnotation true "nginx.ingress.kubernetes.io/force-ssl-redirect" .Spec.Https.ForceRedirect }}

    {{- with .Spec.RateLimit}}
    {{- if .Enabled}}
    {{K8sAnnotation .Rps "nginx.ingress.kubernetes.io/limit-rps" .Rps }}
    {{K8sAnnotation .Rpm "nginx.ingress.kubernetes.io/limit-rpm" .Rpm }}
    {{K8sAnnotation .Connections "nginx.ingress.kubernetes.io/limit-connections" .Connections }}
    {{- end}}
    {{- end}}
spec:
  ingressClassName: ingress-nginx
  tls:
    {{- range $v := $domains }}
    - hosts:
        - {{$v}}
      secretName: {{$v}}-tls
    {{- end }}
  rules:
    {{- range $v := $domains }}
    - host: "*.$v"
      http:
        paths:
          - backend:
              service:
                name: ingress-nginx-{{.Namespace}}-controller
                port:
                  number: 80
    {{- end }}

