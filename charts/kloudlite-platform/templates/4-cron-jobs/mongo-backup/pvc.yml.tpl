{{- $cronName := "mongo-csi-s3-backup" }}

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{$cronName}}
  namespace: {{.Release.Namespace}}
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 200Gi
  storageClassName: {{$cronName}}
