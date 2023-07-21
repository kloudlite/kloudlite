{{- $certManagerName := include "cert-manager.name" . }} 

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$certManagerName}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: jetstack
    url: https://charts.jetstack.io

  chartName: jetstack/cert-manager
  chartVersion: v1.11.0

  valuesYaml: |
    # -- cert-manager args, forcing recursive nameservers used to be google and cloudflare
    # @ignored
    extraArgs:
      - "--dns01-recursive-nameservers-only"
      - "--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"

    tolerations: {{ include "tolerations" . | nindent 6 }}
    nodeSelector: {{ include "node-selector" . | nindent 6 }}

    podLabels: {{ include "pod-labels" . | nindent 6 }}

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
      podLabels: {{ include "pod-labels" . | nindent 8 }}
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
      podLabels: {{ include "pod-labels" . | nindent 8 }}
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
