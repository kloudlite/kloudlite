{{ with . }}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata: {{.Metadata | toJson}}
spec:
  ingressClassName: {{.IngressClassName}}
  {{- if .HttpsEnabled }}
  tls:
    {{- range $v := .NonWildcardDomains }}
    - hosts:
        - {{$v | squote}}
      secretName: {{$v}}-tls
    {{- end}}

    {{- range $v := .WildcardDomains }}
    - hosts:
        - {{$v | squote}}
    {{- end }}
  {{- end}}

  rules:
    {{- range $host, $routes := .Routes }}
    - host: {{$host}}
      http:
        paths:
          {{- range $route := $routes }}
          - pathType: Prefix
            backend:
              service:
                name: {{$route.App}}
                port:
                  number: {{$route.Port}}

            path: {{ if not hasPrefix "/" $route.Path }}/{{end}}{{$route.Path}}
            ({{if hasPrefix "/" $route.Path }}{{substr 1 $x $route.Path}}{{else}}{{$route.Path}}{{end}}.*)

            {{- if $route.Rewrite }}
            path: {{$route.Path}}?(.*)
            {{- else }}
            {{ $x := len $route.Path }}
            path: /({{if hasPrefix "/" $route.Path }}{{substr 1 $x $route.Path}}{{else}}{{$route.Path}}{{end}}.*)
            {{- end}}
          {{- end}}
    {{- end }}
{{- end }}
