{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}

{{- $ownerRefs := get . "owner-refs" | default list }}
{{- $labels := get . "labels" | default dict }}
{{- $annotations := get . "annotations" | default dict }}

{{- $nonWildcardDomains := get . "non-wildcard-domains" }}
{{- $routerDomains := get . "router-domains" }} 
{{- $wildcardDomains := get . "wildcard-domains"}}

{{- $ingressClass := get . "ingress-class" }}
{{- $clusterIssuer := get . "cluster-issuer" }}

{{- $routeToWorkspaceSwitcher := get . "route-to-workspace-switcher" }} 

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
    {{- range $domain := $routerDomains }}
    - host: {{$domain}}
      http:
        paths:
          {{- range $route := $routes }}
          - pathType: Prefix
            backend:
              service:
                name: {{$route.App}}
                port:
                  number: {{$route.Port}}

            {{- if $route.Rewrite }}
            path: {{$route.Path}}?(.*)
            {{- else }}
            {{ $x := len $route.Path }}
            path: /({{if hasPrefix "/" $route.Path }}{{substr 1 $x $route.Path}}{{else}}{{$route.Path}}{{end}}.*)
            {{- end}}
          {{- end}}
    {{- end }}

