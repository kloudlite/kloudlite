apiVersion: batch/v1
kind: Job
metadata:
  name: nats-setup-job
  namespace: {{ $.Release.Namespace }}
  annotations:
    "helm.sh/hook": post-install,post-upgrade
spec:
  template:
    spec:
      containers:
      - name: nats-kv-creator
        image: natsio/nats-box:0.14.1
        command: ["sh"]
        args:
         - -c
         - |
            {{- range $k,$bucketName := .Values.envVars.nats.buckets -}}
            {{- if $.Values.nats.runAsCluster }}
            nats --server nats://nats:4222 kv add {{ $bucketName }} --replicas={{$.Values.nats.replicas}} --storage=file
            {{- else }}
            nats --server nats://nats:4222 kv add {{ $bucketName }} --storage=file
            {{- end }}
            {{- end }}
            {{- range $k,$stream := .Values.envVars.nats.streams -}}
            {{- if $.Values.nats.runAsCluster }}
            nats --server nats://nats:4222 stream add {{ $stream.name }} --replicas={{$.Values.nats.replicas}} --subjects={{ $stream.subjects | squote }} --max-msg-size={{ $stream.maxMsgBytes }} --storage=file --defaults
            {{- else }}
            nats --server nats://nats:4222 stream add {{ $stream.name }}  --replicas={{$.Values.nats.replicas}} --subjects={{ $stream.subjects | squote }} --max-msg-size={{ $stream.maxMsgBytes }} --storage=file --defaults
            {{- end }}
            {{- end }}
      restartPolicy: Never
  backoffLimit: 0