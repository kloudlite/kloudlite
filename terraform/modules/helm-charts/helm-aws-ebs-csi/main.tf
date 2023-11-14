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

resource "ssh_resource" "helm_aws_ebs_csi" {
  host        = var.ssh_params.public_ip
  user        = var.ssh_params.username
  private_key = var.ssh_params.private_key

  timeout     = "1m"
  retry_delay = "5s"

  when = "create"

  pre_commands = [
    "mkdir -p manifests"
  ]

  file {
    content = templatefile("${path.module}/resource.yml", {
      tf_controller_node_selector = var.controller_node_selector
      tf_controller_tolerations   = var.controller_tolerations

      tf_daemonset_node_selector = var.daemonset_node_selector

      tf_storage_classes = local.storage_classes
    })
    destination = "manifests/aws-ebs-csi-driver.yaml"
    permissions = "0666"
  }

  commands = [
    <<EOT
export KUBECTL="sudo k3s kubectl"
$KUBECTL apply -f manifests/aws-ebs-csi-driver.yaml
EOT
  ]
}