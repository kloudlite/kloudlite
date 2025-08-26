{{- $name := .CertSecretNamePrefix }}

{{- $hosts := .Hosts }}
{{- $ingressClass := .IngressClassName }}
{{- $routes := .Routes }} 
{{- $isHttpsEnabled := .HttpsEnabled }} 

ingressClassName: {{$ingressClass}}
{{- if $isHttpsEnabled }}
tls:
  {{- range $host := $hosts }}
  - hosts:
      - {{$host | squote}}
    secretName: {{$name}}-{{$host | replace "." "-"}}-tls
  {{- end}}
{{- end}}

rules:
  {{- range $_, $route := $routes }}
  - host: {{$route.Host}}
    http:
      paths:
        - pathType: ImplementationSpecific
          {{- if $route.Rewrite }}
          path: {{$route.Path}}?(.*)
          {{- else }}
          path: /({{substr 1 (len $route.Path) $route.Path}}.*)
          {{- end}}
          backend:
            service:
              name: {{$route.Service}}
              port:
                number: {{$route.Port}}
  {{- end }}

