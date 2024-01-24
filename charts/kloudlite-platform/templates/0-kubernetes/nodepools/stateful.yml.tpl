{{- $nodepoolName := "stateful" }}
---
apiVersion: clusters.kloudlite.io/v1
kind: NodePool
metadata:
  name: {{$nodepoolName}}
spec:
  maxCount: 3
  minCount: 2

  cloudProvider: aws
  iac:
    stateS3BucketName: {{.Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketName}}
    stateS3BucketRegion: {{.Values.operators.platformOperator.configuration.cluster.IACStateStore.s3BucketRegion}}
    stateS3BucketFilePath: terraform-states/account-{{.Values.global.accountName}}/cluster-{{.Values.global.clusterName}}-{{.Values.global.baseDomain}}/nodepool-{{$nodepoolName}}.tfstate

    cloudProviderAccessKey:
      name: k3s-params
      namespace: kloudlite
      key: accessKey
    cloudProviderSecretKey:
      name: k3s-params
      namespace: kloudlite
      key: secretKey
  aws:
    imageId: "ami-0ec149e1e8b76e957"
    imageSSHUsername: ubuntu
    availabilityZone: ap-south-1a
    nvidiaGpuEnabled: false
    rootVolumeSize: 50
    rootVolumeType: gp3

    poolType: ec2
    ec2Pool:
      instanceType: c6a.large
