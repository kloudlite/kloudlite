apiVersion: clusters.kloudlite.io/v1
kind: NodePool
metadata:
  name: stateless
spec:
  minCount: 2
  maxCount: 2

  nodeTaints: {{.Values.nodepools.stateless.taints | toYaml | nindent 4}}
  nodeLabels: {{.Values.nodepools.stateless.labels | toYaml | nindent 4}}
  
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
          min: "4"
          max: "4"
        memoryPerVcpu:
          min: "2"
          max: "2"
