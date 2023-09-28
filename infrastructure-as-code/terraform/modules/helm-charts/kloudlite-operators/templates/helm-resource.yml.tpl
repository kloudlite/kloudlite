apiVersion: v1
kind: Namespace
metadata:
  name: kloudlite-operators
---
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: operators
  namespace: kloudlite-operators
spec:
  chartRepo:
    name: kloudlite
    url: https://kloudlite.github.io/helm-charts

  chartName: kloudlite/kloudlite-operators
  
  chartVersion: ${kloudlite_release}

  valuesYaml: |+
    preferOperatorsOnMasterNodes: true
    operators:
      helmChartsOperator:
        enabled: false