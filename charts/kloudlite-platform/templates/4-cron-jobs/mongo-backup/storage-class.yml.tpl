{{- $cronName := "mongo-csi-s3-backup" }}
{{- $cronSecretName := "mongo-csi-s3-secret" }}
{{- $ns := .Release.Namespace }}

kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: {{$cronName}}
provisioner: ru.yandex.s3.csi
parameters:
  mounter: geesefs
  # you can set mount options here, for example limit memory cache size (recommended)
  options: "--memory-limit 1000 --dir-mode 0777 --file-mode 0666"
  # to use an existing bucket, specify it here:
  #bucket: some-existing-bucket
  csi.storage.k8s.io/provisioner-secret-name: {{$cronSecretName}}
  csi.storage.k8s.io/provisioner-secret-namespace: {{ $ns }}
  csi.storage.k8s.io/controller-publish-secret-name: {{$cronSecretName}}
  csi.storage.k8s.io/controller-publish-secret-namespace: {{ $ns }}
  csi.storage.k8s.io/node-stage-secret-name: {{$cronSecretName}}
  csi.storage.k8s.io/node-stage-secret-namespace: {{ $ns }}
  csi.storage.k8s.io/node-publish-secret-name: {{$cronSecretName}}
  csi.storage.k8s.io/node-publish-secret-namespace: {{ $ns }}
  bucket: {{.Values.crons.mongoBackup.configuration.bucket}}
