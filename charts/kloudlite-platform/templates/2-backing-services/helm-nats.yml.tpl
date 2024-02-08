{{- $chartName := "nats" }}

---
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartName}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://nats-io.github.io/k8s/helm/charts/
  chartName: nats
  chartVersion: 1.1.5
  jobVars:
    tolerations:
      - operator: Exists
  values:
    global:
      labels:
        kloudlite.io/helmchart: "{{$chartName}}"

    fullnameOverride: {{$chartName}}
    namespaceOverride: {{.Release.Namespace}}

    config:
      cluster:
        enabled: {{.Values.nats.runAsCluster}}
        {{- if .Values.nats.runAsCluster}}
        replicas: {{.Values.nats.replicas}}
        {{- end}}

        routeURLs:
          user: {{.Values.nats.configuration.user}}
          password: {{.Values.nats.configuration.password}}
          useFQDN: true
          k8sClusterDomain: {{.Values.clusterInternalDNS}}

      jetstream:
        enabled: true
        fileStore:
          enabled: true
          dir: /data
          pvc:
            enabled: true
            size: {{.Values.nats.configuration.volumeSize}}
            storageClassName: {{.Values.persistence.storageClasses.xfs}}
            name: {{$chartName}}-jetstream-pvc

{{- if .Values.nats.runAsCluster}}
    podTemplate:
      topologySpreadConstraints:
        kloudlite.io/provider.az:
          maxSkew: 1
          whenUnsatisfiable: DoNotSchedule
          nodeAffinityPolicy: Honor
          nodeTaintsPolicy: Honor
{{- end}}
