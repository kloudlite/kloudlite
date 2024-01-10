{{- $chartOpts :=  .Values.helmCharts.certManager}} 
{{- if $chartOpts.enabled }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://charts.jetstack.io
  chartName: cert-manager
  chartVersion: v1.13.3

  jobVars:
    backOffLimit: 1
    nodeSelector: {{ $chartOpts.nodeSelector | default .Values.defaults.nodeSelector | toYaml | nindent 8}}
    tolerations: {{ $chartOpts.tolerations | default .Values.defaults.tolerations | toYaml | nindent 8}}

  preInstall: |+
    # install cert-manager CRDs
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.crds.yaml

  values:
    # -- cert-manager args, forcing recursive nameservers used to be google and cloudflare
    # @ignored
    extraArgs:
      - "--dns01-recursive-nameservers-only"
      - "--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"

    nodeSelector: {{ $chartOpts.nodeSelector | default .Values.defaults.nodeSelector | toYaml | nindent 8}}
    tolerations: {{ $chartOpts.tolerations | default .Values.defaults.tolerations | toYaml | nindent 8}}

    # -- cert-manager pod affinity
    affinity:
      nodeAffinity: {{ include "preferred-node-affinity-to-masters" . | nindent 8 }}

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

      tolerations: {{ $chartOpts.tolerations | default .Values.defaults.tolerations | toYaml | nindent 8}}

      affinity:
        nodeAffinity: {{ include "preferred-node-affinity-to-masters" . | nindent 10 }}

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

      tolerations: {{ $chartOpts.tolerations | default .Values.defaults.tolerations | toYaml | nindent 8}}

      affinity:
        nodeAffinity: {{ include "preferred-node-affinity-to-masters" . | nindent 10 }}

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

