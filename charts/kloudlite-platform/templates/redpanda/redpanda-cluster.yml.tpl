{{- if .Values.redpandaCluster.create }}
apiVersion: cluster.redpanda.com/v1alpha1
kind: Redpanda
metadata:
  name: redpanda
spec:
  chartRef: {}
  clusterSpec:
    auth:
      sasl:
        enabled: false
        users: []

    tls:
      enabled: false
    external:
      enabled: false

    listeners:
      kafka:
        port: 9092

    statefulset:
      replicas: 1
      nodeSelector: {{.Values.managedServicesNodeSelector | toYaml | nindent 10 }}
      additionalRedpandaCmdFlags:
        - --mode
        - dev-container
      initContainers:
        setDataDirOwnership:
          enabled: true

    resources:
      cpu:
        cores: 1
      memory:
        container:
          min: 1Gi
          max: 1Gi
{{- end }}
