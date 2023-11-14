resource "ssh_resource" "execute_command" {
  host        = var.ssh_params.public_ip
  user        = var.ssh_params.username
  private_key = var.ssh_params.private_key

  timeout     = "${var.timeout_in_minutes}m"
  retry_delay = "5s"

  when = "create"

  commands = [
    <<EOC
function kubectl() {
  command sudo k3s kubectl $@
}
${var.command}
EOC
  ]
}
