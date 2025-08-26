containers:
- name: service-intercept
  image: ghcr.io/kloudlite/hub/socat:latest
  command:
    - sh
    - -c
    - |+
      {{- range $_, $v := .PortMappings }}
      {{- if eq $v.Protocol "TCP" }}
      (socat -dd tcp4-listen:{{$v.ServicePort}},fork,reuseaddr tcp4:{{$.TargetHost}}:{{$v.DevicePort}} 2>&1 | grep -iE --line-buffered 'listening|exiting') &
      pid="$pid $!"
      {{- else if eq $v.Protocol "UDP" }}
      (socat -dd UDP4-LISTEN:{{$v.ServicePort}},fork,reuseaddr UDP4:{{$.TargetHost}}:{{$v.DevicePort}} 2>&1 | grep -iE --line-buffered 'listening|exiting') &
      pid="$pid $!"
      {{- end }}
      {{- end }}

      {{- if not .PortMappings }}
      tail -f /dev/null &
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
