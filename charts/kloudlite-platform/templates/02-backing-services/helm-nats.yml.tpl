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
    tolerations: {{(.Values.nats.tolerations | default .Values.scheduling.stateful.tolerations) | toJson }}
    nodeSelector: {{(.Values.nats.nodeSelector | default .Values.scheduling.stateful.nodeSelector) | toJson }}

  postInstall: |+
    cat <<'EOF' | kubectl apply -f -
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: nats-setup-job-{{ randAlphaNum 5 | lower }}
      namespace: {{ $.Release.Namespace }}
    spec:
      template:
        spec:
          tolerations: {{ (.Values.nats.tolerations | default .Values.scheduling.stateful.tolerations)  | toJson }}
          nodeSelector: {{ (.Values.nats.nodeSelector | default .Values.scheduling.stateful.nodeSelector) | toJson }}
          containers:
          - name: nats-manager
            {{- /* image: natsio/nats-box:0.14.1 */}}
            image: ghcr.io/kloudlite/hub/nats:latest
            command: ["bash"]
            args:
            - -c
            - |+
              echo "creatings NATS KVs"
              {{- range $k,$bucket := .Values.nats.buckets }}
              nats --server {{ include "nats.url" . }} kv add {{ $bucket.name }} {{- if $.Values.nats.runAsCluster}}  --replicas={{$.Values.nats.replicas}} {{- end }} --storage={{$bucket.storage}}
              {{- end }}

              echo "creatings NATS STREAMs"
              {{- range $k,$stream := .Values.nats.streams }}
              params=(
                --server {{include "nats.url" . | squote}}
                --replicas={{$.Values.nats.replicas}}
                --subjects={{ $stream.subjects | squote }}
                --max-msg-size={{ $stream.maxMsgBytes }}
                {{ if $stream.maxMsgsPerSubject }}--max-msgs-per-subject={{$stream.maxMsgsPerSubject}}{{end}}
                {{ if $stream.maxAge }}--max-age={{$stream.maxAge}}{{ end }}
                {{ if $stream.workQueue }}--retention="work"{{ end }}
                --storage=file
                --compression=s2
                --discard=old
                --defaults
              )
              nats stream add "${params[@]}" "{{$stream.name}}"
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
            tolerations: {{(.Values.nats.tolerations | default .Values.scheduling.stateful.tolerations) | toJson }}
            nodeSelector: {{(.Values.nats.nodeSelector | default .Values.scheduling.stateful.nodeSelector) | toJson }}

    podTemplate:
      merge:
        spec:
          tolerations: {{(.Values.nats.tolerations | default .Values.scheduling.stateful.tolerations) | toJson }}
          nodeSelector: {{(.Values.nats.nodeSelector | default .Values.scheduling.stateful.nodeSelector) | toJson }}

      {{- if .Values.nats.runAsCluster}}
      topologySpreadConstraints:
        kloudlite.io/node.name:
          maxSkew: 1
          whenUnsatisfiable: DoNotSchedule
          nodeAffinityPolicy: Honor
          nodeTaintsPolicy: Honor
      {{- end}}
