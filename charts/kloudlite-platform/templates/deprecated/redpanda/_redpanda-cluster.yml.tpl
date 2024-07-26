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
      replicas: 2
      priorityClassName: {{.Values.statefulPriorityClassName}}
      nodeSelector: {{.Values.managedServicesNodeSelector | toYaml | nindent 10 }}
      topologySpreadConstraints:
        maxSkew: 1
        topologyKey: kloudlite.io/provider.az
        whenUnsatisfiable: ScheduleAnyway

      {{- /* additionalRedpandaCmdFlags: */}}
        {{- /* - --check */}}
        {{- /* - "false" */}}
        {{- /* - --overprovisioned */}}
        {{- /* - --unsafe-bypass-fsync */}}
        {{- /* - --overprovisioned */}}
        {{- /* - --mode */}}
        {{- /* - dev-container */}}
      initContainers:
        setDataDirOwnership:
          enabled: true

    resources:
      cpu:
        cores: 1
      memory:
        container:
          min: 2.5Gi
          max: 2.5Gi
          {{- /* min: 1Gi */}}
          {{- /* max: 1Gi */}}
{{- end }}
