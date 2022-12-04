{{- $lokiValues := get . "loki-values" -}}

{{- with $lokiValues }}
{{- /*gotype: operators.kloudlite.io/apis/cluster-setup/v1.LokiValues*/ -}}
loki:
  env:
    - name: AWS_ACCESS_KEY_ID
      value: {{.S3.AwsAccessKeyId}}
    - name: AWS_SECRET_ACCESS_KEY
      value: {{.S3.AwsSecretAccessKey}}
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
        endpoint: {{.S3.Endpoint}}
        bucketnames: {{.S3.BucketName}}
        s3forcepathstyle: true
        insecure: false
        sse_encryption: false
      boltdb_shipper:
        shared_store: s3
        cache_ttl: 24h
grafana:
  enabled: false
  sidecar:
    datasources:
      enabled: true
  image:
    tag: latest
  grafana.ini:
    users:
      default_theme: light
{{- end }}
