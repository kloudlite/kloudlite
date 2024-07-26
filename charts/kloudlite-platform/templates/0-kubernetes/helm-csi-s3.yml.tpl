apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: kl-csi-s3
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://yandex-cloud.github.io/k8s-csi-s3/charts
  chartName: csi-s3
  {{- /* find versions with `helm search repo yandex-s3/csi-s3 --versions` */}}
  chartVersion: "0.41.1"
  values:
    secret:
      create: true

      accessKey: {{ .Values.csiS3.s3.accessKey }}
      secretKey: {{.Values.csiS3.s3.secretKey}}
      endpoint: {{.Values.csiS3.s3.endpoint}}
    
    storageClass:
      create: true
      name: {{.Values.csiS3.storageClass}}

      singleBucket: "{{.Values.csiS3.s3.bucketName}}"
      reclaimPolicy: Retain
