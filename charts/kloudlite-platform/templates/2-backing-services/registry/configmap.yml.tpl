{{- $name := .Release.Name -}}
{{- $namespace := .Release.Namespace -}}

apiVersion: v1
data:
  config.yml: |-
    version: 0.1
    log:
      fields:
        service: registry
    storage:
      delete:
        enabled: true

      {{if .Values.distribution.s3.enabled }}
      s3:
        accesskey: {{ .Values.distribution.s3.accessKey }}
        secretkey: {{ .Values.distribution.s3.secretKey }}
        region: {{ .Values.distribution.s3.region }}
        bucket: {{ .Values.distribution.s3.bucketName }}
        {{ if .Values.distribution.s3.endpoint }}
        regionendpoint: {{ .Values.distribution.s3.endpoint }}
        {{end}}
      {{else}}
      filesystem:
        rootdirectory: /var/lib/registry
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
