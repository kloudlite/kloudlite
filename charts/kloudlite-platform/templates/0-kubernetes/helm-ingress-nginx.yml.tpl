{{- $chartOpts := index .Values.ingressController }}
{{- if $chartOpts.enabled }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://kubernetes.github.io/ingress-nginx

  chartName: ingress-nginx/ingress-nginx
  {{- /* chartVersion: 4.8.0 */}}
  chartVersion: 4.6.0

  values:
    nameOverride: {{$chartOpts.name}}

    rbac:
      create: false

    serviceAccount:
      create: false
      name: {{.Values.global.clusterSvcAccount}}

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

      tolerations: {{ $chartOpts.configuration.tolerations | toYaml | nindent 8 }}
      nodeSelector: {{ $chartOpts.configuration.nodeSelector | toYaml | nindent 8 }}
      affinity: {{$chartOpts.configuration.affinity | toYaml | nindent 8 }}
      {{- end }}

      watchIngressWithoutClass: false
      ingressClassByName: true
      ingressClass: {{$.Values.global.ingressClassName}}
      electionID: {{$.Values.global.ingressClassName}}
      ingressClassResource:
        enabled: true
        name: {{$.Values.global.ingressClassName}}
        controllerValue: "k8s.io/{{$.Values.global.ingressClassName}}"

      {{- if .Values.cloudflareWildCardCert.enabled  }}
      extraArgs:
        default-ssl-certificate: "{{.Release.Namespace}}/{{ .Values.cloudflareWildCardCert.tlsSecretName }}"
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
