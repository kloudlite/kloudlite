locals {
  service_account_name = "kloudlite-admin-sa"
  dir                  = "manifests/kloudlite-agent"
}

resource "ssh_resource" "install-kloudlite-agent" {
  host        = var.ssh_params.host
  user        = var.ssh_params.user
  private_key = var.ssh_params.private_key

  timeout     = "3m"
  retry_delay = "5s"

  when = "create"

  pre_commands = [
    "mkdir -p ${local.dir}"
  ]

  file {
    content = templatefile("${path.module}/templates/admin-service-account.yml", {
      service_account_name      = local.service_account_name
      service_account_namespace = var.release_namespace
    })
    destination = "${local.dir}/admin-service-account.yml"
  }

  file {
    content = templatefile("${path.module}/templates/kloudlite-agent.yml", {
      release_name      = var.release_name
      release_namespace = var.release_namespace
      kloudlite_release = var.kloudlite_release

      service_account_name     = local.service_account_name
      message_office_grpc_addr = var.args.message_office_grpc_addr
      cluster_token            = var.args.cluster_token
      cluster_name             = var.args.cluster_name
      account_name             = var.args.account_name
    })
    destination = "${local.dir}/kloudlite-agent.yml"
  }

  file {
    content = templatefile("${path.module}/templates/helm-charts-controller.yml", {
      controller_name      = "helm-charts-controller"
      controller_namespace = var.release_namespace

      image = "ghcr.io/kloudlite/kloudlite/operator/helm-charts:${var.kloudlite_release}"

      service_account_name = local.service_account_name
      kloudlite_release    = var.kloudlite_release
    })
    destination = "${local.dir}/helm-charts-controller.yml"
  }

  commands = [
    <<-EOC
echo "installing kloudlite agent on tenant cluster"
export KUBECTL="sudo k3s kubectl"

$KUBECTL apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: ${var.release_namespace}
EOF

$KUBECTL apply -f ${local.dir}/admin-service-account.yml
$KUBECTL apply -f ${local.dir}/kloudlite-agent.yml
$KUBECTL apply -f ${local.dir}/helm-charts-controller.yml

EOC
  ]

}