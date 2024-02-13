locals {
  primary_master_node_name = one([
    for node_name, node_cfg in var.master_nodes : node_name
    if node_cfg.role == "primary-master"
  ])
}

resource "null_resource" "variable_validations" {
  lifecycle {
    precondition {
      error_message = "k3s_masters.nodes can/must have only one node with role primary-master"
      condition     = local.primary_master_node_name != null
    }
  }
}

locals {
  master_ssh_params = {
    public_ip   = module.k3s-masters.k3s_primary_public_ip
    username    = var.ssh_username
    private_key = var.ssh_private_key
  }
}

module "constants" {
  source = "../../modules/common/constants"
}

module "k3s-masters" {
  source       = "../../modules/kloudlite/k3s/k3s-master"
  backup_to_s3 = {
    enabled = var.backup_to_s3.enabled

    endpoint      = var.backup_to_s3.endpoint
    bucket_name   = var.backup_to_s3.bucket_name
    bucket_region = var.backup_to_s3.bucket_region
    bucket_folder = var.backup_to_s3.bucket_folder
  }
  cluster_internal_dns_host       = var.cluster_internal_dns_host
  restore_from_latest_s3_snapshot = var.restore_from_latest_snapshot
  master_nodes                    = {
    for k, v in var.master_nodes : k => {
      role : v.role,
      public_ip : v.public_ip,
      node_labels : v.node_labels
    }
  }
  public_dns_host = var.public_dns_host
  ssh_params      = {
    user        = var.ssh_username
    private_key = var.ssh_private_key
  }
  node_taints       = var.taint_master_nodes ? module.constants.master_node_taints : []
  extra_server_args = var.extra_server_args
}

resource "null_resource" "save_kubeconfig" {
  count = length(var.save_kubeconfig_to_path) > 0 ? 1 : 0

  depends_on = [module.k3s-masters.kubeconfig_with_public_host]

  provisioner "local-exec" {
    quiet   = true
    command = "echo '${base64decode(module.k3s-masters.kubeconfig_with_public_host)}' > ${var.save_kubeconfig_to_path} && chmod 600 ${var.save_kubeconfig_to_path}"
  }
}

module "cloudflare-dns" {
  count  = var.cloudflare.enabled ? 1 : 0
  source = "../../modules/cloudflare/dns"

  cloudflare_api_token = var.cloudflare.api_token
  cloudflare_zone_id   = var.cloudflare.zone_id

  DNS_records = concat(
    [
      # A records
      for name, cfg in var.master_nodes :
      {
        record_type = "A"
        domain      = var.cloudflare.domain
        value       = cfg.public_ip
      }
    ],
    [
      # cname records
      {
        record_type = "CNAME",
        domain      = "*.${var.cloudflare.domain}"
        value       = var.cloudflare.domain
      }
    ]
  )
}

module "kloudlite-crds" {
  count             = var.kloudlite_params.install_crds ? 1 : 0
  source            = "../../modules/kloudlite/crds"
  kloudlite_release = var.kloudlite_params.release
  depends_on        = [
    module.k3s-masters.kubeconfig_with_public_host
  ]
  force_apply = var.force_apply_kloudlite_CRDs
  ssh_params  = {
    public_ip   = module.k3s-masters.k3s_primary_public_ip
    username    = var.ssh_username
    private_key = var.ssh_private_key
  }
}

locals {
  kloudlite_namespace = "kloudlite"
}

module "kloudlite-namespace" {
  source     = "../../modules/kloudlite/execute_command_over_ssh"
  depends_on = [
    module.k3s-masters.kubeconfig_with_public_host
  ]
  command    = <<EOF
kubectl apply -f - <<EOF2
apiVersion: v1
kind: Namespace
metadata:
  name: ${local.kloudlite_namespace}
EOF2
EOF
  ssh_params = {
    public_ip   = module.k3s-masters.k3s_primary_public_ip
    username    = var.ssh_username
    private_key = var.ssh_private_key
  }
}

module "kloudlite-k3s-params" {
  source     = "../../modules/kloudlite/execute_command_over_ssh"
  depends_on = [
    module.k3s-masters.kubeconfig_with_public_host
  ]
  pre_command = <<EOF
agent_token=$(sudo cat /var/lib/rancher/k3s/server/agent-token)
server_token=$(sudo cat /var/lib/rancher/k3s/server/node-token)

rm -rf /tmp/k3s-params.yml
cat > /tmp/k3s-params.yml <<EOF2
apiVersion: v1
kind: Secret
metadata:
  name: k3s-params
  namespace: kube-system
stringData:
  k3s_masters_public_dns_host: ${var.public_dns_host}
  k3s_masters_join_token: $server_token
  k3s_agent_join_token: $agent_token

  cloudprovider_name: ${var.cloudprovider_name}
  cloudprovider_region: ${var.cloudprovider_region}
EOF2

kubectl apply -f /tmp/k3s-params.yml
EOF

  command = <<EOF
cat /tmp/k3s-params.yml
EOF

  ssh_params = {
    public_ip   = module.k3s-masters.k3s_primary_public_ip
    username    = var.ssh_username
    private_key = var.ssh_private_key
  }
}

module "fix-k3s-coredns-missing-nodehosts" {
  source     = "../../modules/kloudlite/execute_command_over_ssh"
  depends_on = [
    module.k3s-masters.kubeconfig_with_public_host
  ]
  command = <<EOF
x=$(kubectl get configmap/coredns -n kube-system -o jsonpath='{.data.NodeHosts}')
if [ -z $x ]; then
cat > patch-file.yml <<EOC
metadata:
  annotations:
    kloudlite.io/node-hosts-fixed: "true"
data:
 NodeHosts: ""
EOC

  kubectl patch configmap coredns -n kube-system --patch-file patch-file.yml
fi
EOF

  ssh_params = {
    public_ip   = module.k3s-masters.k3s_primary_public_ip
    username    = var.ssh_username
    private_key = var.ssh_private_key
  }
}

module "nvidia-container-runtime" {
  count      = var.enable_nvidia_gpu_support ? 1 : 0
  source     = "../../modules/nvidia-container-runtime"
  depends_on = [
    module.kloudlite-crds, module.kloudlite-namespace
  ]
  ssh_params        = local.master_ssh_params
  gpu_node_selector = {
    (module.constants.node_labels.node_has_gpu) : "true"
  }
  gpu_node_tolerations = module.constants.gpu_node_tolerations
}

module "kloudlite-agent" {
  count  = var.kloudlite_params.install_agent ? 1 : 0
  source = "../../modules/kloudlite/deployments"
  args   = {
    message_office_grpc_addr = var.kloudlite_params.agent_vars.message_office_grpc_addr
    cluster_token            = var.kloudlite_params.agent_vars.cluster_token

    account_name = var.kloudlite_params.agent_vars.account_name
    cluster_name = var.kloudlite_params.agent_vars.cluster_name
  }
  kloudlite_release = var.kloudlite_params.release
  release_namespace = "kloudlite-tmp"
  ssh_params        = {
    host        = local.master_ssh_params.public_ip
    user        = local.master_ssh_params.username
    private_key = local.master_ssh_params.private_key
  }
}

#module "kloudlite-agent" {
#  count                              = var.kloudlite_params.install_agent ? 1 : 0
#  source                             = "../../modules/kloudlite/helm-kloudlite-agent"
#  kloudlite_account_name             = var.kloudlite_params.agent_vars.account_name
#  kloudlite_cluster_name             = var.kloudlite_params.agent_vars.cluster_name
#  kloudlite_cluster_token            = var.kloudlite_params.agent_vars.cluster_token
#  kloudlite_dns_host                 = var.public_dns_host
#  kloudlite_message_office_grpc_addr = var.kloudlite_params.agent_vars.message_office_grpc_addr
#  kloudlite_release                  = var.kloudlite_params.release
#  ssh_params                         = local.master_ssh_params
#
#  release_name      = "kl-agent"
#  release_namespace = local.kloudlite_namespace
#
#  cloudprovider_name   = var.cloudprovider_name
#  cloudprovider_region = var.cloudprovider_region
#  k3s_agent_join_token = module.k3s-masters.k3s_agent_token
#}

#module "kloudlite-autoscalers" {
#  count             = var.kloudlite_params.install_autoscalers ? 1 : 0
#  source            = "../../modules/kloudlite/helm-kloudlite-autoscalers"
#  depends_on        = [module.kloudlite-crds]
#  kloudlite_release = var.kloudlite_params.release
#  ssh_params        = local.master_ssh_params
#  release_name      = "kl-autoscalers"
#  release_namespace = local.kloudlite_namespace
#}
