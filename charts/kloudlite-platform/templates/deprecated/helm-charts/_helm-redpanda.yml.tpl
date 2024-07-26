apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: redpanda
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: redpanda
    url: https://charts.redpanda.com
  chartName: redpanda/redpanda
  chartVersion: 5.6.52
  values:
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
      initContainers:
        setDataDirOwnership:
          enabled: true

      updateStrategy:
        type: RollingUpdate

      budget:
        maxUnavailable: 1

      podLabels: 
        kloudlite.io/msvc.name: redpanda
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: kloudlite.io/provider.az
          whenUnsatisfiable: DoNotSchedule
          nodeAffinityPolicy: Honor
          nodeTaintsPolicy: Honor
          labelSelector:
            matchLabels:
              kloudlite.io/msvc.name: redpanda

    storage:
      persistentVolume:
        enabled: true
        size: 10Gi
        storageClass: {{.Values.persistence.storageClasses.xfs}}

    resources:
      cpu:
        cores: 1
      memory:
        container:
          min: 2.5Gi
          max: 2.5Gi
