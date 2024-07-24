{{- $cronSecretName := "mongo-csi-s3-secret" }}

apiVersion: v1
kind: Secret
metadata:
  name: {{$cronSecretName}}
  namespace: {{.Release.Namespace}}
stringData:
  accessKeyID: {{.Values.crons.mongoBackup.s3.accessKeyId}}
  secretAccessKey: {{.Values.crons.mongoBackup.s3.secretKey}}
  endpoint: {{.Values.crons.mongoBackup.s3.endpoint}}
