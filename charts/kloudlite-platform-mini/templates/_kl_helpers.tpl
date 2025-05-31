{{- define "image-tag" -}} {{ .Values.kloudliteRelease | default .Chart.AppVersion }} {{- end -}}

{{- define "image-pull-policy" -}}
{{- if .Values.imagePullPolicy -}}
{{- .Values.imagePullPolicy}}
{{- else -}}
{{- if hasSuffix "-nightly" (include "image-tag" .) -}}
{{- "Always" }}
{{- else -}}
{{- "IfNotPresent" }}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "tsc-hostname" -}}
- maxSkew: 1
  topologyKey: kubernetes.io/hostname
  whenUnsatisfiable: DoNotSchedule
  nodeAffinityPolicy: Honor
  nodeTaintsPolicy: Honor
  labelSelector:
    matchLabels: {{ . | toYaml | nindent 6 }}
{{- end -}}

{{- define "default-storage.class" }}
{{- $sc := lookup "storage.k8s.io/v1" "StorageClass" "" "" }}
{{- $items := (index $sc "items") }}

{{- $result := dict }}

{{- range $k, $v := $items }}
  {{- if dig "metadata" "annotations" "storageclass.kubernetes.io/is-default-class" "." $v }}
    {{- $_ := set $result "provisioner" (dig "provisioner" "." $v) }}
    {{- $_ := set $result "reclaimPolicy" (dig "reclaimPolicy" "." $v) }}
  {{- end }}
{{- end }}

{{ $result | toJson }}
{{- end }}

{{- define "common.pod-annotations" }}
{{- merge .Values.podAnnotations (dict "kloudlite.io/helm.last-applied-at" now) | toYaml }}
{{- end }}
p
{{- define "common.pod-labels" }}
{{- .Values.podLabels | default dict | toYaml }}
{{- end }}

{{- define "kloudlite.dns-suffix" -}}
{{.Values.kloudliteDNSSuffix | default (printf "dns.%s" .Values.webHost)}}
{{- end -}}
