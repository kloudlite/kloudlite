{{- define "vector-vm-scrape" -}}
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMServiceScrape
metadata:
  name: vector-vm-scrape
  namespace: {{.Release.Namespace}}
spec:
  endpoints:
  - port: prom-exporter
  namespaceSelector:
    matchNames:
      - {{.Release.Namespace}}
  selector:
    matchLabels:
      app.kubernetes.io/instance: vector
      app.kubernetes.io/name: vector
{{- end -}}
