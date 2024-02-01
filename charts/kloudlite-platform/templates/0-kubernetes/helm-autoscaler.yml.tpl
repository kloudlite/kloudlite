{{- $chartOpts := index .Values.clusterAutoscaler }}
{{- if $chartOpts.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: cluster-autoscaler
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://kloudlite.github.io/helm-charts
  chartName: "kloudlite-autoscalers"
  chartVersion: "{{.Chart.Version}}-nightly"
  jobVars:
    tolerations:
      - operator: Exists
  values:
    defaults:
      imageTag: "v1.0.5-nightly"
      imagePullPolicy: "Always"

    serviceAccount:
      create: true
      nameSuffix: "sa"

    clusterAutoscaler:
      enabled: true
      image:
        repository: "ghcr.io/kloudlite/cluster-autoscaler-amd64"
        tag: kloudlite-v1.0.5-nightly
      nodeSelector:
        node-role.kubernetes.io/master: "true"
      tolerations:
        - operator: Exists
      configuration:
        scaleDownUnneededTime: 1m
{{- end }}
