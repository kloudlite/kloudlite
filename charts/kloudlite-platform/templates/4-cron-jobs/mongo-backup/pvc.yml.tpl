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
      storage: "{{.Values.crons.mongoBackup.configuration.storageSize}}"
  storageClassName: {{.Values.csiS3.storageClass}}
