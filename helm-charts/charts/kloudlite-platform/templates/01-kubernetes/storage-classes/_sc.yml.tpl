{{- range $k, $v := .Values.persistence.storageClasses }}
---
{{- $sc := (lookup "storage.k8s.io/v1" "StorageClass" "" $v) -}}
{{- if $sc }} 
{{$sc | toYaml}}
{{- else }}

{{- $defaultSc := (include "default-storage.class" .| trim | fromJson ) -}}
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{ $v }}
  {{- /* annotations:  */}}
  {{- /*   storageclass.kubernetes.io/is-default-class: "true" */}}
provisioner: {{ get $defaultSc "provisioner" }}
reclaimPolicy: {{  get $defaultSc "reclaimPolicy" }}
volumeBindingMode: WaitForFirstConsumer
parameters:
  fsType: {{$k}}
{{- end }}
{{- end }}
