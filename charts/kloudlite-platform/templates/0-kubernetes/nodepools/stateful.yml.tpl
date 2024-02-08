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
    cloudProviderAccessKey:
      name: k3s-params
      namespace: {{.Release.Namespace}}
      key: accessKey
    cloudProviderSecretKey:
      name: k3s-params
      namespace: {{.Release.Namespace}}
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
