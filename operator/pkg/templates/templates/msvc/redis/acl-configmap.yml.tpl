{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $ownerRefs := get . "owner-refs"  -}}
{{- $aclSecrets := get . "acl-secrets"  | default list -}}

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
data:
  master.conf: |+
    dir /data
    # User-supplied master configuration:
    rename-command FLUSHDB ""
    rename-command FLUSHALL ""
    # End of master configuration

  redis.conf: |+
    {{- range $v := $aclSecrets}}
    {{- if $v }}
    {{- with $v}}
    {{- /*gotype: operators.kloudlite.io/operators/msvc.redis/internal/controllers/types.MresOutput*/ -}}
    {{printf "user %s on ~%s:* +@all -@dangerous +info resetpass >%s" .Username .Prefix .Password | nindent 4 }}
    {{- end}}
    {{- end}}
    {{- end }}

  replica.conf: |+
    dir /data
    slave-read-only yes
    # User-supplied replica configuration:
    rename-command FLUSHDB ""
    rename-command FLUSHALL ""
    # End of replica configuration

