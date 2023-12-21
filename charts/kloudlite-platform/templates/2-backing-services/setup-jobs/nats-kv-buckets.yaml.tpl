{{- range $k,$bucketName := .Values.envVars.natsBuckets -}}
---
apiVersion: batch/v1
kind: Job
metadata:
  name: create-kv-bucket-{{ $bucketName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    "helm.sh/hook": post-install,post-upgrade
spec:
  template:
    spec:
      containers:
      - name: nats-kv-creator
        image: ghcr.io/kloudlite/nats-cli:v1.0.5
        command: ["sh"]
        args:
         - -c
         - |
            {{- if $.Values.nats.runAsCluster }}
            nats --server nats://nats:4222 kv add {{ $bucketName }} --replicas={{$.Values.nats.replicas}} --storage=file
            {{- else }}
            nats --server nats://nats:4222 kv add {{ $bucketName }} --storage=file
            {{- end }}
            sleep 5
      restartPolicy: Never
  backoffLimit: 0
---
{{- end -}}