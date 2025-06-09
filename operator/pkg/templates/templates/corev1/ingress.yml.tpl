{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}

{{- $domains := get . "domains" }}
{{- $wildcardDomains := get . "wildcard-domains"}}

{{- $router := get . "router-ref"}}
{{- $routes := get . "routes" }}
{{- $ownerRefs := get . "owner-refs" | default list }}
{{- $labels := get . "labels" | default dict }}
{{- $annotations := get . "annotations" | default dict }}
{{- $virtualHost := get . "virtual-hostname" |default ""}}

{{- $ingressClass := get . "ingress-class" }}
{{- $clusterIssuer := get . "cluster-issuer" }}

{{- $isInProjectNamespace := get . "is-in-project-namespace" -}}
{{- $bpOverridePort := "80" -}}


{{- with $router}}
{{- /*gotype: github.com/kloudlite/operator/apis/crds/v1.Router */ -}}
{{ $isHttpsEnabled := (or (not .Spec.Https) .Spec.Https.Enabled ) }} 

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  annotations:
    {{- if $annotations}}
    {{ $annotations | toYAML | nindent 4}}
    {{- end}}
  labels: {{ $labels | toYAML | nindent 4}}
  ownerReferences: {{ $ownerRefs | toYAML| nindent 4}}
spec:
  ingressClassName: {{$ingressClass}}
  tls:
  {{- if $isHttpsEnabled }}
    {{- range $v := $domains }}
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
    {{- range $domain := .Spec.Domains }}
    - host: {{$domain}}
      http:
        paths:
          {{- range $route := $routes }}
          - backend:
              service:
                name: {{$route.App | default $route.Lambda}}
                port:
                  number: {{if $isInProjectNamespace}} {{$bpOverridePort}} {{else}}{{$route.Port}}{{end}}

            {{- if $route.Rewrite }}
            path: {{$route.Path}}?(.*)
            {{- else }}
            {{ $x := len $route.Path }}
            path: /({{if hasPrefix "/" $route.Path }}{{substr 1 $x $route.Path}}{{else}}{{$route.Path}}{{end}}.*)
            {{- end}}
            pathType: Prefix
          {{- end}}
    {{- end }}
{{- end}}
