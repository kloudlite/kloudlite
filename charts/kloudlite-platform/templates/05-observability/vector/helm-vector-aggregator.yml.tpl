{{- if .Values.vector.install }}
---
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{ include "vector.name" . }}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://helm.vector.dev
  chartName: vector
  chartVersion: {{ include "vector.chart.version" . | quote }}
  
  jobVars:
    tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 6}}
    nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 6}}

  values:
    global:
      storageClass: {{.Values.persistence.storageClasses.ext4}}

    podAnnotations:
      prometheus.io/scrape: "true"

    replicas: {{.Values.vector.replicas}}
    role: "Stateless-Aggregator"

    tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 6}}
    nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 6}}

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
        {{- /* nats: */}}
        {{- /*   type: nats */}}
        {{- /*   inputs: [vector] */}}
        {{- /*   subject: {{ range .Files.Lines "files/nats-logs-sink-subject.txt" }}  */}}
        {{- /*     {{- . | trim -}} */}}
        {{- /*     {{ end }} */}}
        {{- /*   url: {{ include "nats.url" . }} */}}
        {{- /*   encoding: */}}
        {{- /*       codec: json */}}
        {{- /*       only_fields: */}}
        {{- /*       - message */}}
        {{- /*       - timestamp */}}
        {{- /*       timestamp_format: rfc3339 */}}

        prom_exporter:
          type: prometheus_exporter
          inputs: 
            - vector
          address: 0.0.0.0:9090
          flush_period_secs: 20

        {{- /* loki: */}}
        {{- /*   type: loki */}}
        {{- /*   inputs: */}}
        {{- /*     - vector */}}
        {{- /*   endpoint: http://{{.Values.loki.name}}.{{.Release.Namespace}}:3100 */}}
        {{- /*   encoding: */}}
        {{- /*     codec: json */}}
        {{- /*     only_fields: */}}
        {{- /*       - message */}}
        {{- /*       - timestamp */}}
        {{- /*     timestamp_format: rfc3339 */}}
        {{- /*   labels:  */}}
        {{- /*     {{ range .Files.Lines "files/vector-aggregation-loki-labels.yml" }} */}}
        {{- /*     {{ . -}} */}}
        {{- /*     {{ end }} */}}
        stdout:
          type: console
          inputs: [vector]
          encoding:
            codec: json
{{- end }}
