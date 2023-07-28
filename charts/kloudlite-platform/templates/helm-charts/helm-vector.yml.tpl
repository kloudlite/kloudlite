---
{{- $vectorName := include "vector.name" . }} 

{{- /* -- source: https://vector.dev/docs/reference/configuration/sources/kubernetes_logs/#pod_annotation_fields */}}
{{- define "from-logs.kubernetes.pod_annotations" }}
{{- $labelName := . -}} 
{{- printf "{{- printf \"{{ \" }}" }}
{{- /* should come up, something like 'kubernetes.pod_annotatons."<pod-annotations>"', refer to https://vector.dev/docs/reference/vrl/expressions/#path-example-quoted-path */}}
{{- printf "kubernetes.pod_annotations.\"%s\"" $labelName }}
{{- printf "{{- printf \" }}\" }}" }}
{{ end }}

{{- define "from-logs.kubernetes.namespace_labels" }}
{{- $labelName := . -}} 
{{- printf "{{- printf \"{{ \" }}" }}
{{- /* should come up, something like 'kubernetes.namespace_labels."<label-name>"', refer to https://vector.dev/docs/reference/vrl/expressions/#path-example-quoted-path */}}
{{- printf "kubernetes.namespace_labels.\"%s\"" $labelName }}
{{- printf "{{- printf \" }}\" }}" }}
{{ end }}

{{- /* --- source: https://vector.dev/docs/reference/configuration/sources/kubernetes_logs/#configuration */}}
{{- define "from-logs.kubernetes" }}
{{- $labelName := . -}} 
{{- printf "{{- printf \"{{ \" }}" }}
{{- /* should come up, something like 'kubernetes."<field-name>"', refer to https://vector.dev/docs/reference/vrl/expressions/#path-example-quoted-path */}}
{{- printf "kubernetes.\"%s\"" $labelName }}
{{- printf "{{- printf \" }}\" }}" }}
{{ end }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$vectorName}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: vector
    url: https://helm.vector.dev

  chartName: vector/vector
  chartVersion: 0.23.0

  valuesYaml: |+
    global:
      storageClass: gp2

    podAnnotations:
      prometheus.io/scrape: "true"

    replicas: 1
    role: "Stateless-Aggregator"

    customConfig:
      data_dir: /vector-data-dir
      api:
        enabled: true
        address: 127.0.0.1:8686
        playground: false
      sources:
        vector:
          address: 0.0.0.0:6000
          type: vector
          version: "2"
      sinks:
        prom_exporter:
          type: prometheus_exporter
          inputs: 
            - vector
          address: 0.0.0.0:9090
          flush_period_secs: 20

        loki:
          type: loki
          inputs:
            - vector
          endpoint: http://{{include "loki.name" . }}.{{.Release.Namespace}}:3100
          encoding:
            codec: json
            only_fields:
              {{- /* - stream */}}
              - message
              {{- /* - 'kubernetes.container_image' */}}
              {{- /* - 'kubernetes.container_image_id' */}}
              {{- /* - kubernetes.pod_annotations */}}
              {{- /* - kubernetes.pod_labels */}}
              - timestamp
              {{- /* - timestamp_end */}}
            timestamp_format: rfc3339
          labels: 
            source: vector
            {{- /* INFO: need this after final rendering:  "{{ kubernetes.pod_labels.app }}" */}}
            {{- /* HACK: since custom_config is also tpl rendered, below statements are written for being rendered twice */}}
            kl_account_name: |-
              {{ include "from-logs.kubernetes.namespace_labels" "kloudlite.io/account.name" | trim }}
            kl_cluster_name: |-
              {{ include "from-logs.kubernetes.namespace_labels" "kloudlite.io/cluster.name" | trim }}

            kl_pod_name: |-
              {{ include "from-logs.kubernetes" "pod_name" | trim }}

            kl_container_name: |-
              {{ include "from-logs.kubernetes" "container_name" | trim }}

            kl_container_image: |-
              {{ include "from-logs.kubernetes" "container_image" | trim }}

            kl_container_image_id: |-
              {{ include "from-logs.kubernetes" "container_image_id" | trim }}

            kl_resource_name: |-
              {{ include "from-logs.kubernetes.pod_annotations" "kloudlite.io/resource_name" | trim }}
            kl_resource_namespace: |-
              {{ include "from-logs.kubernetes" "pod_namespace" | trim }}
            kl_resource_type: |-
              {{ include "from-logs.kubernetes.pod_annotations" "kloudlite.io/resource_type" | trim }}
            kl_resource_component: |-
              {{ include "from-logs.kubernetes.pod_annotations" "kloudlite.io/resource_component" | trim }}

            kl_workspace_name: |-
              {{ include "from-logs.kubernetes.pod_annotations" "kloudlite.io/workspace_name" | trim }}
            kl_workspace_target_ns: |-
              {{ include "from-logs.kubernetes.pod_annotations" "kloudlite.io/workspace_target_ns" | trim }}

            kl_project_name: |-
              {{ include "from-logs.kubernetes.pod_annotations" "kloudlite.io/project_name" | trim }}
            kl_project_target_ns: |-
              {{ include "from-logs.kubernetes.pod_annotations" "kloudlite.io/project_target_ns" | trim }}
        stdout:
          type: console
          inputs: [vector]
          encoding:
            codec: json


