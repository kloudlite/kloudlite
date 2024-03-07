module "master-nodes-on-gcp" {
  source                     = "../../../terraform/bundles/gcp/master-nodes"
  machine_type               = var.machine_type
  name_prefix                = var.name_prefix
  nodes                      = var.nodes
  provision_mode             = var.provision_mode
  kloudlite_params           = var.kloudlite_params
  save_ssh_key_to_path       = var.save_ssh_key_to_path
  cloudflare                 = var.cloudflare
  public_dns_host            = var.public_dns_host
  save_kubeconfig_to_path    = var.save_kubeconfig_to_path
  tags                       = var.tags
  label_cloudprovider_region = var.gcp_region
  network                    = "default"
}

