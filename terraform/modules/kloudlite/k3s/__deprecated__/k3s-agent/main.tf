resource "null_resource" "waiting_for_ssh_to_be_active" {
  for_each = {for idx, config in var.agent_nodes : idx => config}
  connection {
    type        = "ssh"
    host        = each.value.public_ip
    user        = each.value.ssh_params.user
    private_key = each.value.ssh_params.private_key
    timeout     = "3m"
  }

  provisioner "remote-exec" {
    inline = [
      <<EOC
echo "node ${each.key}(each.value.public_ip) is ssh-able now"
EOC
    ]
  }
}

resource "ssh_resource" "setup_k3s_on_agents" {
  for_each    = {for idx, config in var.agent_nodes : idx => config}
  host        = each.value.public_ip
  user        = each.value.ssh_params.user
  private_key = each.value.ssh_params.private_key

  #  depends_on = [null_resource.waiting_for_ssh_to_be_active[each.key]]
  depends_on = [null_resource.waiting_for_ssh_to_be_active]

  #  timeout     = "3m"
  #  retry_delay = "2s"

  when = "create"

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
