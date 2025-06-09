{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}

{{- $ownerRefs := get . "owner-refs" | default list }}
{{- $labels := get . "labels" | default dict }}
{{- $annotations := get . "annotations" | default dict }}

{{- $nonWildcardDomains := get . "non-wildcard-domains" }}
{{- $wildcardDomains := get . "wildcard-domains"}}

{{- $ingressClass := get . "ingress-class" }}

{{- $routes := get . "routes" }} 

{{ $isHttpsEnabled := get . "is-https-enabled" }} 

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  annotations: {{ $annotations | toYAML | nindent 4 }}

  labels: {{ $labels | toYAML | nindent 4 }}
  ownerReferences: {{ $ownerRefs | toYAML| nindent 4 }}
spec:
  ingressClassName: {{$ingressClass}}
  {{- if $isHttpsEnabled }}
  tls:
    {{- range $v := $nonWildcardDomains }}
    - hosts:
        - {{$v | squote}}
      secretName: {{$v}}-tls
    {{- end}}

    {{- range $v := $wildcardDomains}}
    - hosts:
        - {{$v | squote}}
    {{- end}}
  {{- end}}

  rules:
    {{- range $_, $route := $routes }}
    - host: {{$route.Host}}
      http:
        paths:
          {{- /* - pathType: Prefix */}}
          - pathType: ImplementationSpecific
            backend:
              service:
                name: {{$route.Service}}
                port:
                  number: {{$route.Port}}

            {{- if $route.Rewrite }}
            path: {{$route.Path}}?(.*)
            {{- else }}
            path: /({{substr 1 (len $route.Path) $route.Path}}.*)
            {{- end}}
    {{- end }}

