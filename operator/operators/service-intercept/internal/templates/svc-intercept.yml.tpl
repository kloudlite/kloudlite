{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}
{{- $labels := get . "labels" | default dict }}
{{- $ownerReferences := get . "owner-references" | default list }}
{{- $deviceHost := get . "device-host" }}

{{- $tcpPortMappings := get . "tcp-port-mappings" | default dict }}
{{- $udpPortMappings := get . "udp-port-mappings" | default dict }}

apiVersion: v1
kind: Pod
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{$labels | toYAML | nindent 4}}
  ownerReferences: {{$ownerReferences | toYAML | nindent 4}}
spec:
  containers:
  - name: app-intercept
    {{- /* image: alpine/socat */}}
    image: ghcr.io/kloudlite/hub/socat:latest
    command:
      - sh
      - -c
      - |+
        {{- range $k, $v := $tcpPortMappings }}
        (socat -dd tcp4-listen:{{$k}},fork,reuseaddr tcp4:{{$deviceHost}}:{{$v}} 2>&1 | grep -iE --line-buffered 'listening|exiting') &
        pid="$pid $!"
        {{- end }}

        {{- range $k, $v := $udpPortMappings }}
        (socat -dd UDP4-LISTEN:{{$k}},fork,reuseaddr UDP4:{{$deviceHost}}:{{$v}} 2>&1 | grep -iE --line-buffered 'listening|exiting') &
        pid="$pid $!"
        {{- end }}

        trap "eval kill -9 $pid || exit 0" EXIT SIGINT SIGTERM
        eval wait $pid
    securityContext:
      capabilities:
        add:
          - NET_BIND_SERVICE
          - SETGID
        drop:
          - all
    resources:
      limits:
        memory: "50Mi"
        cpu: "50m"
      requests:
        memory: "50Mi"
        cpu: "50m"
