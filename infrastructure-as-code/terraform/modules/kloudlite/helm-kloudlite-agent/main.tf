module "constants" {
  source = "../../common/constants"
}

resource "ssh_resource" "apply_kloudlite_crds" {
  host        = var.ssh_params.public_ip
  user        = var.ssh_params.username
  private_key = var.ssh_params.private_key

  timeout     = "1m"
  retry_delay = "5s"

  when = "create"

  pre_commands = [
    "mkdir -p manifests"
  ]

  triggers = {
    always_run = timestamp()
  }

  #  file {
  #    content = templatefile("${path.module}/templates/helm-charts-agent.yml.tpl", {
  #      release_name                       = var.release_name
  #      release_namespace                  = var.release_namespace
  #      kloudlite_release                  = var.kloudlite_release
  #      kloudlite_account_name             = var.kloudlite_account_name
  #      kloudlite_cluster_name             = var.kloudlite_cluster_name
  #      kloudlite_cluster_token            = var.kloudlite_cluster_token
  #      kloudlite_message_office_grpc_addr = var.kloudlite_message_office_grpc_addr
  #      kloudlite_dns_host                 = var.kloudlite_dns_host
  #      helm_job_tolerations               = module.constants.master_node_tolerations
  #    })
  #    destination = "manifests/helm-charts-kloudlite-agent.yml"
  #  }

  file {
    content = templatefile("${path.module}/templates/helm-install-agent-with-job.yml.tpl", {
      svc_account_name      = "install-agent-sa"
      svc_account_namespace = var.release_namespace

      release_name      = var.release_name
      release_namespace = var.release_namespace

      kloudlite_release                  = var.kloudlite_release
      kloudlite_account_name             = var.kloudlite_account_name
      kloudlite_cluster_name             = var.kloudlite_cluster_name
      kloudlite_cluster_token            = var.kloudlite_cluster_token
      kloudlite_message_office_grpc_addr = var.kloudlite_message_office_grpc_addr

      cloudprovider_name      = var.cloudprovider_name
      cloudprovider_region    = var.cloudprovider_region
      k3s_agent_join_token    = var.k3s_agent_join_token
      k3s_masters_public_host = var.kloudlite_dns_host

      kloudlite_dns_host   = var.kloudlite_dns_host
      helm_job_tolerations = module.constants.master_node_tolerations
    })
    destination = "manifests/helm-charts-kloudlite-agent.yml"
  }

  commands = [
    <<EOC
export KUBECTL="sudo k3s kubectl"
$KUBECTL apply -f manifests/helm-charts-kloudlite-agent.yml
EOC
  ]
}
