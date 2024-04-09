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
    tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 6 }}
    nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 6 }}
  values:
    global:
      labels:
        kloudlite.io/helmchart: "{{$chartName}}"

    fullnameOverride: {{$chartName}}
    namespaceOverride: {{.Release.Namespace}}

    {{- if .Values.nats.runAsCluster }}
    container:
      env:
        # different from k8s units, suffix must be B, KiB, MiB, GiB, or TiB
        # should be ~90% of memory limit
        GOMEMLIMIT: 2700MiB
      merge:
        # recommended limit is at least 2 CPU cores and 8Gi Memory for production JetStream clusters
        resources:
          requests:
            cpu: "1"
            memory: 3Gi
          limits:
            cpu: "1"
            memory: 3Gi
    {{- end }}

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
          k8sClusterDomain: {{.Values.global.clusterInternalDNS}}

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

    natsBox:
      enabled: true
      podTemplate:
        merge:
          spec:
            tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 12 }}
            nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 12 }}

    podTemplate:
      merge:
        spec:
          tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 12}}
          nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 12}}
{{- if .Values.nats.runAsCluster}}
      topologySpreadConstraints:
        kloudlite.io/provider.az:
          maxSkew: 1
          whenUnsatisfiable: DoNotSchedule
          nodeAffinityPolicy: Honor
          nodeTaintsPolicy: Honor
        kloudlite.io/node.name:
          maxSkew: 1
          whenUnsatisfiable: DoNotSchedule
          nodeAffinityPolicy: Honor
          nodeTaintsPolicy: Honor
{{- end}}
