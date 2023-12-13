{{- $chartOpts := index .Values.helmCharts "kube-prometheus" }} 

{{- if $chartOpts.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: bitnami
    url: https://charts.bitnami.com/bitnami

  chartName: bitnami/kube-prometheus

  chartVersion: 8.15.1

  values:
    global:
      storageClass: {{.Values.persistence.storageClasses.ext4}}

    nameOverride: {{$chartOpts.name}}
    fullnameOverride: {{$chartOpts.name}}

    operator:
      enabled: true
      service:
        kubeletService:
          enabled: false
      
    prometheus:
      enabled: true
      image:
        registry: docker.io
        repository: bitnami/prometheus
        tag: 2.45.0-debian-11-r2
        digest: ""

      enableAdminApi: true
      retention: 10d
      disableCompaction: false
      walCompression: false
      persistence:
        enabled: true
        size: {{$chartOpts.configuration.prometheus.volumeSize}}
      paused: false

      nodeSelector: {{$chartOpts.configuration.prometheus.nodeSelector | toYaml |nindent 10}}

      additionalScrapeConfigs:
        enabled: true
        type: internal
        internal:
          jobList:
            - job_name: "kubernetes-pods"
              kubernetes_sd_configs:
                - role: pod

              relabel_configs:
                # Example relabel to scrape only pods that have
                # "example.io/should_be_scraped = true" annotation.
                - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
                  action: keep
                  regex: true

                # - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
                #   action: keep
                #   regex: vector

                - action: labelmap
                  regex: __meta_kubernetes_pod_label_(.+)
                - source_labels: [__meta_kubernetes_namespace]
                  action: replace
                  target_label: namespace
                - source_labels: [__meta_kubernetes_pod_name]
                  action: replace
                  target_label: pod

    alertmanager:
      enabled: true
      image:
        registry: docker.io
        repository: bitnami/alertmanager
        tag: 0.25.0-debian-11-r65
        digest: ""

      persistence:
        enabled: true
        size: {{$chartOpts.configuration.prometheus.volumeSize}}
      paused: false

      nodeSelector: {{$chartOpts.configuration.prometheus.nodeSelector | toYaml |nindent 10}}
    
    exporters:
      node-exporter:
        enabled: false
      kube-state-metrics:
        enabled: false

    kubelet:
      enabled: false

    blackboxExporter:
      enabled: false

    kubeApiServer:
      enabled: false
    kubeControllerManager:
      enabled: false
    kubeScheduler:
      enabled: false
    coreDns:
      enabled: false
    kubeProxy:
      enabled: false
{{- end }}

