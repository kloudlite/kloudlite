{{- define "vector-vm-scrape" -}}
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMPodScrape
metadata:
  name: vector-aggregator-scrape
  namespace: {{.Release.Namespace}}
spec:
  podMetricsEndpoints:
  - port: prom-exporter
  namespaceSelector:
    matchNames:
      - {{.Release.Namespace}}
  selector:
    matchLabels:
      app.kubernetes.io/instance: vector
      app.kubernetes.io/name: vector
{{- end -}}
