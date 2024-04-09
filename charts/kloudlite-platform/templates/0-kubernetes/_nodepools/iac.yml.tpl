{{- if (and .Values.nodepools.enabled .Values.nodepools.iac.enabled) }}
apiVersion: clusters.kloudlite.io/v1
kind: NodePool
metadata:
  name: iac
spec:
  minCount: {{.Values.nodepools.iac.min}}
  maxCount: {{.Values.nodepools.iac.max}}

  nodeTaints: {{.Values.nodepools.iac.taints | toYaml | nindent 4}}
  nodeLabels: {{.Values.nodepools.iac.labels | toYaml | nindent 4}}

  cloudProvider: aws
  aws:
    vpcId: {{.Values.nodepools.iac.aws.vpcId}}
    vpcSubnetId: {{.Values.nodepools.iac.aws.vpcSubnetId}}
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
{{- end }}
