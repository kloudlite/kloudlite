{{- if .Values.nginxIngress.install }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{ include "nginx-ingress.name" . }}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://kubernetes.github.io/ingress-nginx
  chartName: ingress-nginx
  chartVersion: {{ include "nginx-ingress.chart.version" . }}

  values:
    nameOverride: {{ include "nginx-ingress.name" . }}

    rbac:
      create: true

    serviceAccount:
      create: true
      name: {{ include "nginx-ingress.name" . }}-sa

    controller:
      # -- ingress nginx controller configuration
      kind: Deployment
      service:
        type: LoadBalancer

      watchIngressWithoutClass: false
      ingressClassByName: true
      ingressClass: {{.Values.nginxIngress.ingressClass.name}}
      electionID: {{.Values.nginxIngress.ingressClass.name}}
      ingressClassResource:
        enabled: true
        name: {{.Values.nginxIngress.ingressClass.name}}
        controllerValue: "k8s.io/{{.Values.nginxIngress.ingressClass.name}}"

      {{- if .Values.nginxIngress.wildcardCert.enabled  }}
      extraArgs:
        default-ssl-certificate: '{{.Values.nginxIngress.wildcardCert.secretNamespace}}/{{.Values.nginxIngress.wildcardCert.secretName}}'
      {{- end }}

      podLabels: {{ .Values.podLabels | toYaml | nindent 8 }}

      resources:
        requests:
          cpu: 100m
          memory: 200Mi

      admissionWebhooks:
        enabled: false
        failurePolicy: Ignore
{{- end }}
