{{- $ingressNginxName := include "ingress-nginx.name" . }} 

{{- $subchartOpts := index .Values.subcharts "ingress-nginx" }} 

{{- $ingressClassName := $subchartOpts.ingressClassName }} 

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$ingressNginxName}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: ingress-nginx
    url: https://kubernetes.github.io/ingress-nginx

  chartName: ingress-nginx/ingress-nginx
  chartVersion: 4.6.0

  valuesYaml: |+
    nameOverride: {{$ingressNginxName}}

    rbac:
      create: false

    serviceAccount:
      create: false
      name: {{.Values.clusterSvcAccount}}

    controller:
      # -- ingress nginx controller configuration
      {{- if (eq $subchartOpts.controllerKind "Deployment") }}
      {{- printf `
      kind: Deployment
      service:
        type: LoadBalancer
      `}}
      {{- end }}

      {{- if (eq $subchartOpts.controllerKind "DaemonSet") }}
      {{- printf `
      kind: DaemonSet
      service:
        type: "ClusterIP"

      hostNetwork: true
      hostPort:
        enabled: true
        ports:
          http: 80
          https: 443
          healthz: 10254

      dnsPolicy: ClusterFirstWithHostNet
      nodeSelector:
        node-role.kubernetes.io/control-plane: "true"
      `}}
      {{- end }}

      watchIngressWithoutClass: false
      ingressClassByName: true
      ingressClass: {{$ingressClassName}}
      electionID: {{$ingressClassName}}
      ingressClassResource:
        enabled: true
        name: {{$ingressClassName}}
        controllerValue: "k8s.io/{{$ingressClassName}}"

      {{- if .Values.cloudflareWildCardCert.create  }}
      {{- printf `
      extraArgs:
        default-ssl-certificate: "%s/%s"
      ` .Release.Namespace (include "cloudflare-wildcard-certificate.secret-name" .) }} 
      {{- end }}

      podLabels: {{ include "pod-labels" . | nindent 8}}

      resources:
        requests:
          cpu: 100m
          memory: 200Mi

      admissionWebhooks:
        enabled: false
        failurePolicy: Ignore
