{{- $name := get . "name" -}}
{{- $fsTypes := get . "fs-types" -}}
{{- $labels := get . "labels"  -}}

{{- range $fsType := $fsTypes }}
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{$name}}-{{$fsType}}
  labels: {{$labels | toYAML | nindent 4}}
provisioner: ebs.csi.aws.com
parameters:
  csi.storage.k8s.io/fstype: {{$fsType}}
reclaimPolicy: Delete
{{/*volumeBindingMode: WaitForFirstConsumer*/}}
volumeBindingMode: Immediate
allowVolumeExpansion: true
{{- end }}
