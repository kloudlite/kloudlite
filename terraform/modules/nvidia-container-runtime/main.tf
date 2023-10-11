locals {
  nvidia_script = templatefile("${path.module}/scripts/nvidia-gpu-post-k3s-start-without-helm.sh", {
    TF_GPU_NODES_SELECTOR = jsonencode(var.gpu_nodes_selector)
  })
  destination_path = "./manifests/nvidia-gpu-post-k3s-start.sh"
}

resource "ssh_resource" "setup_nvidia_gpu_on_node" {
  host        = var.ssh_params.public_ip
  user        = var.ssh_params.user
  private_key = var.ssh_params.private_key

  timeout     = "1m"
  retry_delay = "2s"

  when = "create"

  pre_commands = [
    "mkdir -p ./manifests"
  ]

  file {
    content     = local.nvidia_script
    destination = local.destination_path
    permissions = "0755"
  }

  triggers = {
    script_content = local.nvidia_script
  }

  commands = [
    <<EOT
chmod +x ${local.destination_path}
sudo bash ${local.destination_path} 2>&1 | tee nvidia-gpu-script.log
EOT
  ]
}
