resource "ssh_resource" "disable_ssh_on_nodes" {
  for_each    = {for idx, node in var.nodes_config : idx => node}
  host        = each.value.public_ip
  user        = each.value.ssh_params.user
  private_key = each.value.ssh_params.private_key

  commands = [
    <<EOC
if [ "${each.value.disable_ssh}" == "true" ]; then
  sudo systemctl disable sshd.service
  sudo systemctl stop sshd.service
  sudo rm -f ~/.ssh/authorized_keys
fi
EOC
  ]
}
