{{- $cronSecretName := "nats-csi-s3-secret" }}

apiVersion: v1
kind: Secret
metadata:
  name: {{$cronSecretName}}
  namespace: {{.Release.Namespace}}
stringData:
  accessKeyID: {{.Values.crons.natsBackup.s3.accessKeyId}}
  secretAccessKey: {{.Values.crons.natsBackup.s3.secretKey}}
  endpoint: {{.Values.crons.natsBackup.s3.endpoint}}
