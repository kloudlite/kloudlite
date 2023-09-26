{{- $lokiName := include "loki.name" . }} 

{{- $subchartOpts := index .Values.subcharts "loki-stack"  }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$lokiName}}
  namespace: {{.Release.Namespace}}
spec:
  chartRepo:
    name: grafana
    url: https://grafana.github.io/helm-charts

  chartName: grafana/loki-stack
  chartVersion: 2.9.10

  valuesYaml: |+
    loki:
      enabled: true
      env:
        {{- if $subchartOpts.s3credentials.awsAccessKeyId }}
        - name: AWS_ACCESS_KEY_ID
          value: {{$subchartOpts.s3credentials.awsAccessKeyId | squote}}
        {{- end }}

        {{- if $subchartOpts.s3credentials.awsSecretAccessKey }}
        - name: AWS_SECRET_ACCESS_KEY
          value: {{$subchartOpts.s3credentials.awsSecretAccessKey | squote}}
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
            {{/* # s3: s3://us-west-2/bluelightco-loki1729 */}}
            {{/* # s3: s3://sgp1.digitaloceanspaces.com/plaxspace */}}
            {{/* # endpoint: sgp1.digitaloceanspaces.com */}}
            {{/* # endpoint: s3.ap-south-1.amazonaws.com */}}
            {{/* s3: s3://ap-south-1/kloudlite-logs */}}
            s3: s3://{{$subchartOpts.s3credentials.region}}/{{$subchartOpts.s3credentials.bucketName}}
            s3forcepathstyle: true
            insecure: false
            sse_encryption: false
          boltdb_shipper:
            shared_store: s3
            cache_ttl: 24h

    promtail:
      enabled: false

