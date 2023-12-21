{{- $chartOpts := index .Values.loki }}
{{- if $chartOpts.enabled }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$chartOpts.name}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: grafana
    url: https://grafana.github.io/helm-charts

  chartName: grafana/loki-stack
  chartVersion: 2.9.10

  values:
    loki:
      enabled: true
      priorityClassName: {{.Values.global.statefulPriorityClassName}}
      env:
        {{- if $chartOpts.configuration.s3credentials.awsAccessKeyId }}
        - name: AWS_ACCESS_KEY_ID
          value: {{$chartOpts.configuration.s3credentials.awsAccessKeyId | squote}}
        {{- end }}

        {{- if $chartOpts.configuration.s3credentials.awsSecretAccessKey }}
        - name: AWS_SECRET_ACCESS_KEY
          value: {{$chartOpts.configuration.s3credentials.awsSecretAccessKey | squote}}
        {{- end }}
      config:
        schema_config:
          configs:
            - from: 2021-05-12
              store: boltdb-shipper
              object_store: s3
              schema: v11
              index:
                prefix: loki_index_
                period: 24h
        storage_config:
          aws:
            s3: s3://{{$chartOpts.configuration.s3credentials.region}}/{{$chartOpts.configuration.s3credentials.bucketName}}
            s3forcepathstyle: true
            insecure: false
            sse_encryption: false
          boltdb_shipper:
            shared_store: s3
            cache_ttl: 24h

    promtail:
      enabled: false

{{- end }}

