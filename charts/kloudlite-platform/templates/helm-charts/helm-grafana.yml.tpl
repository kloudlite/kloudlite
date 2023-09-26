{{- $grafanaName := include "grafana.name" . }} 

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$grafanaName}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: bitnami
    url: https://charts.bitnami.com/bitnami

  chartName: bitnami/grafana
  chartVersion: 9.0.1

  valuesYaml: |+
    global:
      storageClass: {{.Values.persistence.storageClasses.ext4}}

    nameOverride: {{$grafanaName}}
    fullnameOverride: {{$grafanaName}}

    persistence:
      enabled: true
      size: 2Gi

