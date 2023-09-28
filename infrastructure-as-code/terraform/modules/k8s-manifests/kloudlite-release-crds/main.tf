resource "ssh_resource" "apply_kloudlite_crds" {
  host        = var.ssh_params.public_ip
  user        = var.ssh_params.username
  private_key = var.ssh_params.private_key

  timeout     = "1m"
  retry_delay = "5s"

  when = "create"

  commands = [
    <<EOC
export KUBECTL="sudo k3s kubectl"
$KUBECTL apply -f https://github.com/kloudlite/helm-charts/releases/download/${var.kloudlite_release}/crds-all.yml
EOC
  ]
}