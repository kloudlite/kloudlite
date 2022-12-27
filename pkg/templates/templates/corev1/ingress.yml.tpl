{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}

{{- $domains := get . "domains" }}
{{- $wildcardDomains := get . "wildcard-domains"}}

{{- $router := get . "router-ref"}}
{{- $routes := get . "routes" }}
{{- $ownerRefs := get . "owner-refs"}}
{{- $labels := get . "labels"}}
{{- $annotations := get . "annotations"}}
{{- $virtualHost := get . "virtual-hostname" |default ""}}

{{- $ingressClass := get . "ingress-class" }}
{{- $clusterIssuer := get . "cluster-issuer" }}

{{- $certIngressClass := get . "cert-ingress-class"  -}}
{{- $globalIngressClass := get . "global-ingress-class"  -}}

{{- $isBlueprint := get . "is-blueprint" -}}
{{- $bpOverridePort := "80" -}}

{{- with $router}}
{{- /*gotype: operators.kloudlite.io/apis/crds/v1.Router */ -}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  annotations:
    {{ K8sAnnotation true "cert-manager.io/cluster-issuer" $clusterIssuer }}

    {{- $bodySize := printf "%dm" .Spec.MaxBodySizeInMB}}
    {{K8sAnnotation .Spec.MaxBodySizeInMB "nginx.ingress.kubernetes.io/proxy-body-size" $bodySize }}
    nginx.ingress.kubernetes.io/ssl-redirect: {{.Spec.Https.Enabled | squote }}
    nginx.ingress.kubernetes.io/force-ssl-redirect: {{.Spec.Https.ForceRedirect | squote}}
{{/*    nginx.ingress.kubernetes.io/from-to-www-redirect: "true"*/}}


    {{- with .Spec.RateLimit}}
    {{- if .Enabled}}
    {{K8sAnnotation .Rps "nginx.ingress.kubernetes.io/limit-rps" .Rps }}
    {{K8sAnnotation .Rpm "nginx.ingress.kubernetes.io/limit-rpm" .Rpm }}
    {{K8sAnnotation .Connections "nginx.ingress.kubernetes.io/limit-connections" .Connections }}
    {{- end}}
    {{- end}}

{{/*    {{K8sAnnotation true "nginx.ingress.kubernetes.io/rewrite-target" "$1" }}*/}}
    {{K8sAnnotation true "nginx.ingress.kubernetes.io/rewrite-target" "/$1" }}
    {{K8sAnnotation $virtualHost "nginx.ingress.kubernetes.io/upstream-vhost" $virtualHost}}
    {{K8sAnnotation true "nginx.ingress.kubernetes.io/preserve-trailing-slash" "true"}}

{{/*    basic auth*/}}
    {{- if .Spec.BasicAuth.Enabled }}
    nginx.ingress.kubernetes.io/auth-type: basic
    nginx.ingress.kubernetes.io/auth-secret: {{.Spec.BasicAuth.SecretName }}
    nginx.ingress.kubernetes.io/auth-realm: 'Route is protected by basic authentication'
    {{- end }}

    {{- if $annotations}}
    {{ $annotations | toYAML | nindent 4}}
    {{- end}}
  labels: {{ $labels | toYAML | nindent 4}}
  ownerReferences: {{ $ownerRefs | toYAML| nindent 4}}
spec:
  ingressClassName: {{$ingressClass}}
  tls:
    {{- range $v := $domains }}
    - hosts:
        - {{$v | squote}}
      secretName: {{$v}}-tls
    {{- end}}
    {{- range $v := $wildcardDomains}}
    - hosts:
        - {{$v | squote}}
    {{- end}}
  rules:
    {{- range $domain := .Spec.Domains }}
    - host: www.{{$domain}}
      http:
        paths:
          {{- range $route := $routes }}
          - backend:
              service:
                name: {{$route.App | default $route.Lambda}}
                port:
                  number: {{if $isBlueprint}} {{$bpOverridePort}} {{else}}{{$route.Port}}{{end}}

            {{- if $route.Rewrite }}
            path: {{$route.Path}}?(.*)
            {{- else}}
            {{ $x := len $route.Path }}
{{/*            {{if not (gt $x 1)}}*/}}
{{/*            path: "/(.*)"*/}}
{{/*            {{else}}*/}}
            path: /({{if hasPrefix "/" $route.Path }}{{substr 1 $x $route.Path}}{{else}}{{$route.Path}}{{end}}.*)
{{/*            {{end}}*/}}
            {{- end}}
            pathType: Prefix
          {{- end}}
    - host: {{$domain}}
      http:
        paths:
          {{- range $route := $routes }}
          - backend:
              service:
                name: {{$route.App | default $route.Lambda}}
                port:
                  number: {{if $isBlueprint}} {{$bpOverridePort}} {{else}}{{$route.Port}}{{end}}

            {{- if $route.Rewrite }}
            path: {{$route.Path}}?(.*)
            {{- else}}
            {{ $x := len $route.Path }}
{{/*            {{if not (gt $x 1)}}*/}}
{{/*            path: "/(.*)"*/}}
{{/*            {{else}}*/}}
            path: /({{if hasPrefix "/" $route.Path }}{{substr 1 $x $route.Path}}{{else}}{{$route.Path}}{{end}}.*)
{{/*            {{end}}*/}}
            {{- end}}
            pathType: Prefix
          {{- end}}
    {{- end }}
  {{- end}}
