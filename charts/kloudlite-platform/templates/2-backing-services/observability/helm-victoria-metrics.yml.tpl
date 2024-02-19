{{- if .Values.victoriaMetrics.enabled }}

{{- $chartVersion := "0.18.11" }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{.Values.victoriaMetrics.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://victoriametrics.github.io/helm-charts/
  chartName: victoria-metrics-k8s-stack

  chartVersion: {{$chartVersion}}

  jobVars:
    tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 6}}
    nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 6}}

  preInstall: |+
    curl -L0 'https://raw.githubusercontent.com/VictoriaMetrics/helm-charts/victoria-metrics-k8s-stack-{{$chartVersion}}/charts/victoria-metrics-k8s-stack/charts/crds/crds/crd.yaml' | kubectl apply -f -

  postInstall: |+
    kubectl apply -f - <<EOF
    {{ include "vector-vm-scrape" .  | nindent 6 }}
    EOF

  values:
    fullnameOverride: {{.Values.victoriaMetrics.name}}
    grafana:
      enabled: false

    crds:
      enabled: false

    victoria-metrics-operator:
      createCRD: false
      tolerations: {{.Values.nodepools.stateless.tolerations | toYaml | nindent 8 }}
      nodeSelector: {{.Values.nodepools.stateless.labels | toYaml | nindent 8 }}

    kube-state-metrics:
      enabled: true
      tolerations: {{.Values.nodepools.stateless.tolerations | toYaml | nindent 8 }}
      nodeSelector: {{.Values.nodepools.stateless.labels | toYaml | nindent 8 }}

    prometheus-node-exporter:
      enabled: true
      tolerations:
        - operator: Exists

    vmsingle:
      enabled: false

    vmagent:
      enabled: true
      spec:
        tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 10 }}
        nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 10 }}

    vmalert:
      enabled: true
      spec:
        tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 10 }}
        nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 10 }}

    alertmanager:
      enabled: true
      spec:
        tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 10 }}
        nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 10 }}

    vmcluster:
      enabled: true
      spec:
        retentionPeriod: "1d"

        vmstorage:
          tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 12 }}
          nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 12 }}

          storage:
            volumeClaimTemplate:
              spec:
                storageClassName: "{{.Values.persistence.storageClasses.ext4}}"
                resources:
                  requests:
                    storage: {{.Values.victoriaMetrics.configuration.vmcluster.volumeSize}}

        vmselect:
          tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 12 }}
          nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 12 }}

          storage:
            volumeClaimTemplate:
              spec:
                storageClassName: "{{.Values.persistence.storageClasses.ext4}}"
                resources:
                  requests:
                    storage: {{.Values.victoriaMetrics.configuration.vmselect.volumeSize}}

        vminsert:
          tolerations: {{.Values.nodepools.stateful.tolerations | toYaml | nindent 12 }}
          nodeSelector: {{.Values.nodepools.stateful.labels | toYaml | nindent 12 }}

    nameOverride: {{.Values.victoriaMetrics.name}}
    fullnameOverride: {{.Values.victoriaMetrics.name}}
{{- end -}}
