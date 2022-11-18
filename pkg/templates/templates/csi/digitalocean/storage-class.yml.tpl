{{- $name := get . "name" -}}
{{- $fsTypes := get . "fs-types" -}}
{{- $labels := get . "labels"  -}}
{{- $ownerRefs := get . "owner-refs"  -}}


{{- range $fsType := $fsTypes }}
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{$name}}-{{$fsType}}
  labels: {{$labels | toYAML | nindent 4}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
provisioner: dobs.csi.digitalocean.com
parameters:
  fsType: {{$fsType}}
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
{{- end }}
