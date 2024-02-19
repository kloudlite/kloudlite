apiVersion: clusters.kloudlite.io/v1
kind: NodePool
metadata:
  name: iac
spec:
  minCount: 2
  maxCount: 10

  nodeTaints: {{.Values.nodepools.iac.taints | toYaml | nindent 4}}
  nodeLabels: {{.Values.nodepools.iac.labels | toYaml | nindent 4}}

  cloudProvider: aws
  aws:
    imageId: "ami-0ec149e1e8b76e957"
    imageSSHUsername: ubuntu
    availabilityZone: ap-south-1a
    nvidiaGpuEnabled: false
    rootVolumeSize: 80
    rootVolumeType: gp3

    poolType: spot
    spotPool:
      spotFleetTaggingRoleName: "kloudlite-platform-role"
      cpuNode:
        vcpu:
          min: "2"
          max: "2"
        memoryPerVcpu:
          min: "2"
          max: "2"
