locals {
  storage_classes = [
    for sc_name, sc_config in var.storage_classes : {
      name : sc_name
      labels : tomap({
        "kloudlite.io/installed-by" : "kloudlite-iac"
      })
      volumeBindingMode : "WaitForFirstConsumer"
      reclaimPolicy : "Retain"
      parameters : tomap({
        encrypted : "false"
        type : sc_config.volume_type
        fsType : sc_config.fs_type
      })
    }
  ]
}

resource "helm_release" "aws_ebs_csi_driver" {
  name = "aws-ebs-csi-driver"

  repository = "https://kubernetes-sigs.github.io/aws-ebs-csi-driver"
  chart      = "aws-ebs-csi-driver"

  version          = "2.22.0"
  namespace        = "kube-system"
  create_namespace = false

  values = [
    <<EOT
customLabels:
  kloudlite.io/installed-by: "kloudlite-iac"
storageClasses:
${yamlencode(local.storage_classes)}
    EOT
  ]
}