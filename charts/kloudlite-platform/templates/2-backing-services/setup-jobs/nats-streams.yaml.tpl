{{- range $k,$stream := .Values.envVars.natsStreams -}}
---
apiVersion: batch/v1
kind: Job
metadata:
  name: create-stream-{{ $stream.name }}
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
            nats --server nats://nats:4222 stream add {{ $stream.name }} --replicas={{$.Values.nats.replicas}} --subjects={{ $stream.subjects | squote }} --max-msg-size={{ $stream.maxMsgBytes }} --storage=file --defaults
            {{- else }}
            nats --server nats://nats:4222 stream add {{ $stream.name }}  --replicas={{$.Values.nats.replicas}} --subjects={{ $stream.subjects | squote }} --max-msg-size={{ $stream.maxMsgBytes }} --storage=file --defaults
            {{- end }}
            sleep 5
      restartPolicy: Never
  backoffLimit: 0
---
{{- end -}}