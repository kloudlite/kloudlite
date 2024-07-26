apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{.Values.crons.natsBackup.name}}
  namespace: {{.Release.Namespace}}
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 50Gi
  storageClassName: {{.Values.csiS3.storageClass}}
