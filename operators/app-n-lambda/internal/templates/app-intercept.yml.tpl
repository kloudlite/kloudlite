{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}
{{- $labels := get . "labels" | default dict }}
{{- $ownerReferences := get . "owner-references" | default list }}
{{- $deviceHost := get . "device-host" }}
{{- $portMappings := get . "port-mappings" | default dict }}

apiVersion: v1
kind: Pod
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{$labels | toYAML | nindent 4}}
  ownerReferences: {{$ownerReferences | toYAML | nindent 4}}
spec:
  dnsPolicy: Default
  containers:
  - name: app-intercept
    image: alpine/socat
    command:
      - sh
      - -c
      - |+
        {{- range $k, $v := $portMappings }}
        {{- /* (socat -dd tcp4-listen:{{$k}},fork,reuseaddr tcp4:{{$deviceHost}}:{{$v}} 2>&1 | grep -iE 'accepting|exiting') & */}}
        (socat -dd tcp4-listen:{{$k}},fork,reuseaddr tcp4:{{$deviceHost}}:{{$v}}) &
        pid="$pid $!"
        {{- end }}
        {{- /* (socat -dd tcp4-listen:80,fork,reuseaddr tcp4:first-device.device.local:17171 2>&1 | grep -iE 'accepting|exiting') & */}}

        trap "cleanup" EXIT INT TERM
        trap "eval kill -9 $pid || exit 0" EXIT SIGINT SIGTERM
        eval wait $pid
    securityContext:
      capabilities:
        drop:
          - ALL
    resources:
      limits:
        memory: "20Mi"
        cpu: "20m"
