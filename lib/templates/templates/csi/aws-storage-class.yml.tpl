{{- $name := get . "name" -}}
{{- $fsTypes := get . "fs-types" -}}
{{- $driverName := get . "driver-name"  -}}


{{- range $fsType := $fsTypes }}
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{$name}}-{{$fsType}}
provisioner: {{$driverName}}.ebs.csi.aws.com
parameters:
  csi.storage.k8s.io/fstype: $fsType
reclaimPolicy: Delete
{{/*volumeBindingMode: WaitForFirstConsumer*/}}
volumeBindingMode: Immediate
allowVolumeExpansion: true
{{- end }}
