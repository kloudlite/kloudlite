apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{.Values.crons.mongoBackup.name}}
  namespace: {{.Release.Namespace}}
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: {{.Values.csiS3.storageClass}}
