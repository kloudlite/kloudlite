#resource "null_resource" "setup_k3s_on_agents" {
#  for_each = {for idx, config in var.agent_nodes : idx => config}
#  connection {
#    type        = "ssh"
#    user        = each.value.ssh_params.user
#    host        = each.value.public_ip
#    private_key = each.value.ssh_params.private_key
#  }
#
#  provisioner "remote-exec" {
#    inline = [
#      <<-EOC
#      cat >runner-config.yml <<EOF2
#      runAs: agent
#      agent:
#        publicIP: ${each.value.public_ip}
#        serverIP: ${var.k3s_server_dns_hostname}
#        token: ${var.k3s_token}
#        labels: ${jsonencode(each.value.node_labels)}
#        nodeName: ${each.key}
#      EOF2
#
#      sudo ln -sf $PWD/runner-config.yml /runner-config.yml
#      if [ "${var.use_cloudflare_nameserver}" = "true" ]; then
#        lineNo=$(sudo cat /etc/resolv.conf -n | grep "nameserver" | awk '{print $1}')
#        sudo sed -i "$lineNo i nameserver 1.1.1.1" /etc/resolv.conf
#      fi
#      EOC
#    ]
#  }
#}

resource "ssh_resource" "setup_k3s_on_agents" {
  for_each    = {for idx, config in var.agent_nodes : idx => config}
  host        = each.value.public_ip
  user        = each.value.ssh_params.user
  private_key = each.value.ssh_params.private_key

  commands = [
    <<EOC
cat >runner-config.yml <<EOF2
runAs: agent
agent:
  publicIP: ${each.value.public_ip}
  serverIP: ${var.k3s_server_dns_hostname}
  token: ${var.k3s_token}
  labels: ${jsonencode(each.value.node_labels)}
  nodeName: ${each.key}
EOF2

sudo ln -sf $PWD/runner-config.yml /runner-config.yml
if [ "${var.use_cloudflare_nameserver}" = "true" ]; then
  lineNo=$(sudo cat /etc/resolv.conf -n | grep "nameserver" | awk '{print $1}')
  sudo sed -i "$lineNo i nameserver 1.1.1.1" /etc/resolv.conf
fi
EOC
  ]
}

resource "ssh_resource" "setup_nvidia_gpu_on_node" {
  for_each    = {for idx, config in var.agent_nodes : idx => config if config.is_nvidia_gpu_node == true}
  host        = each.value.public_ip
  user        = each.value.ssh_params.user
  private_key = each.value.ssh_params.private_key

  timeout     = "10s"
  retry_delay = "2s"

  when = "create"

  pre_commands = [
    "mkdir -p manifests"
  ]

  file {
    content     = file("${path.module}/../scripts/nvidia-gpu-post-k3s-start.sh")
    destination = "manifests/nvidia-gpu-post-k3s-start.sh"
  }

  commands = [
    <<EOT
sudo bash manifests/nvidia-gpu-post-k3s-start.sh
EOT
  ]
}
