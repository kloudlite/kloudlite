apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  ame: mongo-ha
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: bitnami
    url: https://charts.bitnami.com/bitnami

  chartName: bitnami/mongodb
  chartVersion: 14.3.1

  values:
    global:
      storageClass: sc-xfs
    architecture: replicaset
    image:
      registry: docker.io
      repository: bitnami/mongodb
      tag: 7.0.2-debian-11-r0

    fullnameOverride: mongo-ha
    replicaCount: 3
    replicaSetName: rs
    replicaSetHostNames: true

    podLabels:
      kloudlite.io/msvc.name: mongo-ha

    directoryPerDB: true

    persistence:
      enabled: true
      size: 1Gi

    auth:
      enabled: true
    
    topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kloudlite.io/provider.az
        whenUnsatisfiable: DoNotSchedule
        nodeAffinityPolicy: Honor
        nodeTaintsPolicy: Honor
        labelSelector:
          matchLabels:
            kloudlite.io/msvc.name: mongo-ha

    volumePermissions:
      enabled: true
