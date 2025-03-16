{{- $releaseName :=  include "nats.name" . }}

---
apiVersion: plugin-helm-chart.kloudlite.github.com/v1
kind: HelmChart
metadata:
  name: {{$releaseName}}
  namespace: {{.Release.Namespace}}
spec:
  chart:
    url: https://nats-io.github.io/k8s/helm/charts/ 
    name: nats
    version: 1.1.5
  jobVars:
    tolerations: {{(.Values.nats.tolerations | default .Values.scheduling.stateful.tolerations) | toYaml | nindent 6 }}
    nodeSelector: {{(.Values.nats.nodeSelector | default .Values.scheduling.stateful.nodeSelector) | toYaml | nindent 6 }}

  postInstall: |+
    cat <<EOF | kubectl apply -f -
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
                {{if $stream.maxMsgsPerSubject }} --max-msgs-per-subject={{$stream.maxMsgsPerSubject}} {{end}} \
                --storage=file \
                {{ if $stream.maxAge }} --max-age={{$stream.maxAge}} {{ end }} \
                {{ if $stream.workQueue }} --retention="work" {{ end }} \
                --compression=s2 \
                --discard=old \
                --defaults
              {{- end }}
          restartPolicy: Never
      backoffLimit: 0
    EOF

  helmValues:
    global:
      labels:
        kloudlite.io/helmchart: "{{$releaseName}}"

    fullnameOverride: {{$releaseName}}
    namespaceOverride: {{.Release.Namespace}}

    {{- if .Values.nats.runAsCluster }}
    container:
      env:
        # different from k8s units, suffix must be B, KiB, MiB, GiB, or TiB
        # should be ~90% of memory limit
        GOMEMLIMIT: 2700MiB
      merge:
        # recommended limit is at least 2 CPU cores and 8Gi Memory for production JetStream clusters
        resources:
          requests:
            cpu: "1"
            memory: 3Gi
          limits:
            cpu: "1"
            memory: 3Gi
    {{- end }}

    config:
      cluster:
        enabled: {{.Values.nats.runAsCluster}}
        {{- if .Values.nats.runAsCluster}}
        replicas: {{.Values.nats.replicas}}
        {{- end}}

        routeURLs:
          {{- /* user: {{.Values.nats.configuration.user}} */}}
          {{- /* password: {{.Values.nats.configuration.password}} */}}
          useFQDN: true
          k8sClusterDomain: {{.Values.clusterInternalDNS}}

      jetstream:
        enabled: true
        fileStore:
          enabled: true
          dir: /data
          pvc:
            enabled: true
            size: {{.Values.nats.volumeSize}}
            storageClassName: {{.Values.persistence.storageClasses.xfs}}
            name: {{$releaseName}}-jetstream-pvc

    natsBox:
      enabled: true
      podTemplate:
        merge:
          spec:
            tolerations: {{(.Values.nats.tolerations | default .Values.scheduling.stateful.tolerations) | toYaml | nindent 14 }}
            nodeSelector: {{(.Values.nats.nodeSelector | default .Values.scheduling.stateful.nodeSelector) | toYaml | nindent 14 }}

    podTemplate:
      merge:
        spec:
          tolerations: {{(.Values.nats.tolerations | default .Values.scheduling.stateful.tolerations) | toYaml | nindent 12}}
          nodeSelector: {{(.Values.nats.nodeSelector | default .Values.scheduling.stateful.nodeSelector) | toYaml | nindent 12}}

      {{- if .Values.nats.runAsCluster}}
      topologySpreadConstraints:
        {{- /* kloudlite.io/provider.az: */}}
        {{- /*   maxSkew: 1 */}}
        {{- /*   whenUnsatisfiable: DoNotSchedule */}}
        {{- /*   nodeAffinityPolicy: Honor */}}
        {{- /*   nodeTaintsPolicy: Honor */}}
        kloudlite.io/node.name:
          maxSkew: 1
          whenUnsatisfiable: DoNotSchedule
          nodeAffinityPolicy: Honor
          nodeTaintsPolicy: Honor
      {{- end}}
