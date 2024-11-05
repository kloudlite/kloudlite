resource "ssh_resource" "kloudlite_operators" {
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
    content = templatefile("${path.module}/templates/helm-charts-operator-deploy.yml.tpl", {
      deployment_name      = "helm-charts-operator"
      deployment_namespace = var.release_namespace

      image             = "ghcr.io/kloudlite/operators/helm-charts:${var.kloudlite_release}"
      image_pull_policy = "Always"

      svc_account_name      = "helm-charts-svc-account"
      svc_account_namespace = var.release_namespace
    })
    destination = "manifests/helm-charts-operator-deploy.yml"
    permissions = "0666"
  }

  file {
    content = templatefile("${path.module}/templates/helm-release-kloudlite-operators.yml.tpl", {
      release_name      = var.release_name
      release_namespace = var.release_namespace
      kloudlite_release = var.kloudlite_release
    })
    destination = "manifests/kloudlite-operators.yml"
    permissions = "0666"
  }

  commands = [
    <<EOT
export KUBECTL="sudo k3s kubectl"
$KUBECTL apply -f manifests/helm-charts-operator-deploy.yml
$KUBECTL apply -f manifests/kloudlite-operators.yml
EOT
  ]
}
