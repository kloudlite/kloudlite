{{- $chartName := "nats" }}

---
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartName}}
  namespace: kloudlite
spec:
  chartRepo:
    name: nats
    url: https://nats-io.github.io/k8s/helm/charts/
  chartName: nats/nats
  chartVersion: 1.1.5
  jobVars:
    tolerations:
      - operator: Exists
  values:
    global:
      labels:
        kloudlite.io/helmchart: "{{$chartName}}"

    fullnameOverride: {{$chartName}}
    namespaceOverride: kloudlite

    config:
      cluster:
        enabled: true
        replicas: 3

        routeURLs:
          user: sample
          password: sample
          useFQDN: true
          k8sClusterDomain: cluster.local

      jetstream:
        enabled: true

        fileStore:
          enabled: true
          dir: /data

          pvc:
            enabled: true
            size: 10Gi
            storageClassName: {{.Values.persistence.storageClasses.xfs}}
            name: {{$chartName}}-jetstream-pvc

    podTemplate:
      topologySpreadConstraints: 
        kloudlite.io/provider.az:
          maxSkew: 1
          whenUnsatisfiable: DoNotSchedule
          nodeAffinityPolicy: Honor
          nodeTaintsPolicy: Honor
