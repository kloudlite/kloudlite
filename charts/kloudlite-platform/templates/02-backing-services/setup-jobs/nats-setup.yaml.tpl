apiVersion: batch/v1
kind: Job
metadata:
  name: nats-setup-job-{{ randAlphaNum 5 | lower }}
  namespace: {{ $.Release.Namespace }}
spec:
  template:
    spec:
      tolerations: {{ (.Values.nats.tolerations | default .Values.scheduling.stateful.tolerations)  |toYaml| nindent 8 }}
      nodeSelector: {{ (.Values.nats.nodeSelector | default .Values.scheduling.stateful.nodeSelector) | toYaml | nindent 8}}
      containers:
      - name: nats-manager
        image: natsio/nats-box:0.14.1
        command: ["sh"]
        args:
         - -c
         - |+
           echo "creatings NATS KVs"
           {{- range $k,$bucket := .Values.nats.buckets }}
           nats --server {{ include "nats.url" . }} kv add {{ $bucket.name }} {{- if $.Values.nats.runAsCluster}}  --replicas={{$.Values.nats.replicas}} {{- end }} --storage={{$bucket.storage}}
           {{- end }}

           echo "creatings NATS STREAMs"
           {{- range $k,$stream := .Values.nats.streams }}
           nats --server {{include "nats.url" .}} stream add {{ $stream.name }} \
             --replicas={{$.Values.nats.replicas}} \
             --subjects={{ $stream.subjects | squote }} \
             --max-msg-size={{ $stream.maxMsgBytes }} \
             {{ if $stream.maxMsgsPerSubject }} --max-msgs-per-subject={{$stream.maxMsgsPerSubject}} {{ end }} \
             --storage=file \
             {{ if $stream.maxAge }} --max-age={{$stream.maxAge}} {{ end }} \
             {{ if $stream.workQueue }} --retention="work" {{ end }} \
             --compression=s2 \
             --discard=old \
             --defaults
           {{- end }}
      restartPolicy: Never
  backoffLimit: 0
