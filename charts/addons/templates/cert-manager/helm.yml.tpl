{{- if .Values.common.certManager.enabled }}

{{- $tolerations := .Values.common.certManager.configuration.tolerations | toYaml }}
{{- $nodeSelector := .Values.common.certManager.configuration.nodeSelector | toYaml }}

{{- $certManagerChartVersion := "v1.13.3" }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: "cert-manager"
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://charts.jetstack.io
  chartName: cert-manager
  chartVersion: {{$certManagerChartVersion}}

  jobVars:
    backOffLimit: 1
    nodeSelector: {{ $nodeSelector | nindent 6 }}
    tolerations: {{ $tolerations | nindent 6 }}

  preInstall: |+
    echo installing cert-manager CRDs
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/{{$certManagerChartVersion}}/cert-manager.crds.yaml
    echo installed cert-manager CRDs

  values:
    # -- cert-manager args, forcing recursive nameservers used to be google and cloudflare
    # @ignored
    extraArgs:
      - "--dns01-recursive-nameservers-only"
      - "--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"

    nodeSelector: {{ .Values.common.certManager.configuration.nodeSelector | toYaml | nindent 6 }}
    tolerations: {{ .Values.common.certManager.configuration.tolerations | toYaml | nindent 6 }}

    # -- cert-manager pod affinity
    podLabels: {{ .Values.podLabels | toYaml | nindent 6 }}

    startupapicheck:
      # -- whether to enable startupapicheck, disabling it by default as it unnecessarily increases chart installation time
      enabled: false

    resources:
      # -- resource limits for cert-manager controller pods
      limits:
        # -- cpu limit for cert-manager controller pods
        cpu: 80m
        # -- memory limit for cert-manager controller pods
        memory: 120Mi
      requests:
        # -- cpu request for cert-manager controller pods
        cpu: 40m
        # -- memory request for cert-manager controller pods
        memory: 120Mi

    webhook:
      podLabels: {{ .Values.podLabels | toYaml | nindent 8 }}

      nodeSelector: {{ $nodeSelector | nindent 8 }}
      tolerations: {{ $tolerations | nindent 8 }}

      # -- resource limits for cert-manager webhook pods
      resources:
        # -- resource limits for cert-manager webhook pods
        limits:
          # -- cpu limit for cert-manager webhook pods
          cpu: 60m
          # -- memory limit for cert-manager webhook pods
          memory: 60Mi
        requests:
          # -- cpu limit for cert-manager webhook pods
          cpu: 30m
          # -- memory limit for cert-manager webhook pods
          memory: 60Mi

    cainjector:
      podLabels: {{ .Values.podLabels | toYaml | nindent 8 }}

      nodeSelector: {{ $nodeSelector | nindent 8 }}
      tolerations: {{ $tolerations | nindent 8 }}

      # -- resource limits for cert-manager cainjector pods
      resources:
        # -- resource limits for cert-manager webhook pods
        limits:
          # -- cpu limit for cert-manager cainjector pods
          cpu: 120m
          # -- memory limit for cert-manager cainjector pods
          memory: 200Mi
        requests:
          # -- cpu requests for cert-manager cainjector pods
          cpu: 80m
          # -- memory requests for cert-manager cainjector pods
          memory: 200Mi

{{- end }}
