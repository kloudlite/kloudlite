{{- $deviceHost := .DeviceHost }}
{{- $tcpPortMappings := .TCPPortMappings }}
{{- $udpPortMappings := .UDPPortMappings }}

containers:
- name: app-intercept
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
