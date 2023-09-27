{{- $chartOpts := index .Values.helmCharts "ingress-nginx" }} 
{{- if $chartOpts.enabled }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: ingress-nginx
    url: https://kubernetes.github.io/ingress-nginx

  chartName: ingress-nginx/ingress-nginx
  chartVersion: 4.8.0

  valuesYaml: |+
    nameOverride: {{$chartOpts.name}}

    rbac:
      create: false

    serviceAccount:
      create: false
      name: {{.Values.clusterSvcAccount}}

    controller:
      # -- ingress nginx controller configuration
      {{- if (eq $chartOpts.configuration.controllerKind "Deployment") }}
      kind: Deployment
      service:
        type: LoadBalancer
      {{- end }}

      {{- if (eq $chartOpts.configuration.controllerKind "DaemonSet") }}
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
      affinity: {{$chartOpts.configuration.affinity | toYaml | nindent 8 }}
      {{- end }}

      watchIngressWithoutClass: false
      ingressClassByName: true
      ingressClass: {{$chartOpts.configuration.ingressClassName}}
      electionID: {{$chartOpts.configuration.ingressClassName}}
      ingressClassResource:
        enabled: true
        name: {{$chartOpts.configuration.ingressClassName}}
        controllerValue: "k8s.io/{{$chartOpts.configuration.ingressClassName}}"

      {{- if .Values.cloudflareWildCardCert.create  }}
      extraArgs:
        default-ssl-certificate: "{{.Release.Namespace}}/{{ include "cloudflare-wildcard-certificate.secret-name" . }}"
      {{- end }}

      podLabels: {{ include "pod-labels" . | nindent 8 }}

      resources:
        requests:
          cpu: 100m
          memory: 200Mi

      admissionWebhooks:
        enabled: false
        failurePolicy: Ignore

{{- end }}
