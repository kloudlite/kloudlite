module "master-nodes" {
  source                        = "../../../terraform/bundles/gcp/master-nodes"
  machine_type                  = var.machine_type
  name_prefix                   = var.name_prefix
  nodes                         = var.nodes
  provision_mode                = var.provision_mode
  kloudlite_params              = var.kloudlite_params
  save_ssh_key_to_path          = var.save_ssh_key_to_path
  cloudflare                    = var.cloudflare
  public_dns_host               = var.public_dns_host
  save_kubeconfig_to_path       = var.save_kubeconfig_to_path
  tags                          = var.tags
  label_cloudprovider_region    = var.gcp_region
  network                       = var.network
  use_as_longhorn_storage_nodes = var.use_as_longhorn_storage_nodes
}

module "worker-nodes" {
  source = "../../../terraform/bundles/gcp/worker-nodes"

  depends_on = [module.master-nodes.kubeconfig]

  for_each = {for name, cfg in var.nodepools : name => cfg}

  allow_incoming_http_traffic = false
  availability_zone           = each.value.availability_zone
  k3s_extra_agent_args        = each.value.k3s_extra_agent_args
  k3s_join_token              = module.master-nodes.k3s_agent_token
  k3s_server_public_dns_host  = module.master-nodes.k3s_public_dns_host
  kloudlite_release           = var.kloudlite_params.release
  machine_type                = each.value.machine_type
  name_prefix                 = "${var.name_prefix}-${each.key}"
  network                     = var.network
  nodes                       = each.value.nodes
  node_labels                 = each.value.node_labels
  provision_mode              = each.value.provision_mode
  nodepool_name               = each.key
  bootvolume_type             = each.value.bootvolume_type
  bootvolume_size             = each.value.bootvolume_size
  additional_disk             = each.value.additional_disk
}

