{{- $name := .Release.Name -}}
{{- $namespace := .Release.Namespace -}}

apiVersion: v1
data:
  gcp-credentials.json: |+
    {{.Values.distribution.storage.gcs.keyfileJson}}
  config.yml: |-
    version: 0.1
    log:
      level: debug
      fields:
        service: registry
    storage:
      delete:
        enabled: true

      {{- if eq .Values.distribution.storage.driver "gcs"}}
      gcs:
        bucket: {{.Values.distribution.storage.gcs.bucket}}
        keyfile: /etc/docker/registry/gcp-credentials.json
      {{- end }}

      {{if eq .Values.distribution.storage.driver "s3" }}
      s3:
        accesskey: {{ .Values.distribution.storage.s3.accessKey }}
        secretkey: {{ .Values.distribution.storage.s3.secretKey }}
        region: {{ .Values.distribution.storage.s3.region }}
        bucket: {{ .Values.distribution.storage.s3.bucketName }}
        {{ if .Values.distribution.storage.s3.endpoint }}
        regionendpoint: {{ .Values.distribution.storage.s3.endpoint }}
        {{end}}
        v4Auth: false
        secure: true
      {{- /* {{else}} */}}
      {{- /* filesystem: */}}
      {{- /*   rootdirectory: /var/lib/registry */}}
      {{end}}

    http:
      addr: :5000
      secret: {{ .Values.distribution.secret }}
      headers:
        X-Content-Type-Options: [nosniff]
    health:
      storagedriver:
        enabled: true
        interval: 10s
        threshold: 3

    notifications:
      events:
        includereferences: true
      endpoints:
        - name: alistener
          disabled: false
          url: http://container-registry-api:4000/events
          timeout: 1s
          threshold: 10
          backoff: 1s
          ignoredmediatypes:
            - application/octet-stream

kind: ConfigMap
metadata:
  name: {{ $name }}
