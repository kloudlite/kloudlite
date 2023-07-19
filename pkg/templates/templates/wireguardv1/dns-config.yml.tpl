{{- $rewriteRules := get . "rewrite-rules"}}
{{- $dnsIp := get . "dns-ip"}}
{{- $devices := get . "devices"}}
{{- $namespace := get . "namespace"}}

apiVersion: v1
data:
  devices: |
    {{ $devices | toJson }}
  Corefile: |
    .:53 {
        errors
        health
        ready

        {{ $rewriteRules }}

        forward . {{ $dnsIp }}
        cache 30
        loop
        reload
        loadbalance
    }
    import /etc/coredns/custom/*.server
kind: ConfigMap
metadata:
  name: coredns
  namespace: {{ $namespace }}
