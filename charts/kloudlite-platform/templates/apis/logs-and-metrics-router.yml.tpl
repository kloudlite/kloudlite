---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.observabilityApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ (index .Values.helmCharts "ingress-nginx").configuration.ingressClassName }}

  domains:
    - "{{.Values.routers.observabilityApi.name}}.{{.Values.baseDomain}}"
  https:
    enabled: true
    clusterIssuer: {{.Values.clusterIssuer.name}}
    forceRedirect: true
  routes:
    - app: {{.Values.apps.consoleApi.name}}
      path: /
      port: 9100
---

{{/*apiVersion: networking.k8s.io/v1*/}}
{{/*kind: Ingress*/}}
{{/*metadata:*/}}
{{/*  annotations:*/}}
{{/*    cert-manager.io/cluster-issuer: cluster-issuer*/}}
{{/*    kloudlite.io/group-version-kind: networking.k8s.io/v1, Kind=Ingress*/}}
{{/*    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"*/}}
{{/*    nginx.ingress.kubernetes.io/preserve-trailing-slash: "true"*/}}
{{/*    nginx.ingress.kubernetes.io/proxy-body-size: 50m*/}}
{{/*    nginx.ingress.kubernetes.io/rewrite-target: /$1*/}}
{{/*    nginx.kubernetes.io/ssl-redirect: "true"*/}}
{{/*  name: observability*/}}
{{/*  namespace: kl-core*/}}
{{/*spec:*/}}
{{/*  ingressClassName: ingress-nginx*/}}
{{/*  rules:*/}}
{{/*  - host: observability.{{.Values.baseDomain}}*/}}
{{/*    http:*/}}
{{/*      paths:*/}}
{{/*      - backend:*/}}
{{/*          service:*/}}
{{/*            name: console-api*/}}
{{/*            port:*/}}
{{/*              number: 9100*/}}
{{/*        path: /observability/(.*)*/}}
{{/*        pathType: Prefix*/}}
{{/*  tls:*/}}
{{/*  - hosts:*/}}
{{/*    - '*.{{.Values.baseDomain}}'*/}}
{{/*---*/}}
