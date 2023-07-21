{{- $vectorName := include "vector.name" . }} 

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$vectorName}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: vector
    url: https://helm.vector.dev

  chartName: vector/vector
  chartVersion: 0.23.0

  valuesYaml: |+
    global:
      storageClass: gp2

    podAnnotations:
      prometheus.io/scrape: "true"

    replicas: 1
    role: "Stateless-Aggregator"

    customConfig:
      data_dir: /vector-data-dir
      api:
        enabled: true
        address: 127.0.0.1:8686
        playground: false
      sources:
        vector:
          address: 0.0.0.0:6000
          type: vector
          version: "2"
      sinks:
        prom_exporter:
          type: prometheus_exporter
          inputs: 
            - vector
          address: 0.0.0.0:9090
          flush_period_secs: 20

        loki:
          type: loki
          inputs:
            - vector
          endpoint: http://{{include "loki.name" . }}.{{.Release.Namespace}}:3100
          encoding:
            codec: logfmt
          labels: 
            source: vector
            {{/* INFO: need this after final rendering:  "{{ kubernetes.pod_labels.app }}" */}}
            {{/* HACK: since custom_config is also tpl rendered, below statements are written for being rendered twice */}}
            kl_app: |-
              {{ "{{ print \"{{\" }}" }} {{print "kubernetes.pod_labels.app" }} {{ "{{ print \"}}\" }}" }}

        stdout:
          type: console
          inputs: [vector]
          encoding:
            codec: json


