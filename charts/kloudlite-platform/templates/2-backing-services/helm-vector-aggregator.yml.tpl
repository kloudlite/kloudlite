{{- $chartOpts := .Values.vector }}
{{- if $chartOpts.enabled }}

---
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://helm.vector.dev
  chartName: vector
  chartVersion: 0.23.0
  
  values:
    global:
      storageClass: {{.Values.persistence.storageClasses.ext4}}

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
        nats:
          type: nats
          inputs: [vector]
          {{- /* subject: {{ range .Files.Get "files/nats-logs-sink-subject.txt" | nindent 13 | trim | splitList "\n" }} */}}
          {{- /*   {{- . | trim -}} */}}
          {{- /*   {{ end}} */}}
          subject: |+
            {{ range .Files.Lines "files/nats-logs-sink-subject.txt" }} 
            {{- . | trim -}}
            {{ end }}
          url: {{.Values.envVars.nats.url}}
          encoding:
              codec: json
              only_fields:
              - message
              - timestamp
              timestamp_format: rfc3339

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
          endpoint: http://{{.Values.loki.name}}.{{.Release.Namespace}}:3100
          encoding:
            codec: json
            only_fields:
              - message
              - timestamp
            timestamp_format: rfc3339
          labels: 
            {{ range .Files.Lines "files/vector-aggregation-loki-labels.yml" }}
            {{ . -}}
            {{ end }}
        stdout:
          type: console
          inputs: [vector]
          encoding:
            codec: json

  {{- end }}
