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
            echo "creatings NATS KVs"
            {{- range $k,$bucket := .Values.envVars.nats.buckets }}
            nats --server nats://nats:4222 kv add {{ $bucket.name }} {{- if $.Values.nats.runAsCluster}}  --replicas={{$.Values.nats.replicas}} {{- end }} --storage={{$bucket.storage}}
            {{- end }}

            echo "creatings NATS STREAMs"
            {{- range $k,$stream := .Values.envVars.nats.streams }}
            nats --server nats://nats:4222 stream add {{ $stream.name }} --replicas={{$.Values.nats.replicas}} --subjects={{ $stream.subjects | squote }} --max-msg-size={{ $stream.maxMsgBytes }} --storage=file --defaults
            {{- end }}
      restartPolicy: Never
  backoffLimit: 0
