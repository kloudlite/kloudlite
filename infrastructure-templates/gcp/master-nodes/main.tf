resource "random_id" "name_suffix" {
  keepers = {
    # Generate a new id each time we switch to a new AMI name prefix
    name_prefix = var.name_prefix
  }

  byte_length = 4
}

module "master-nodes-on-gcp" {
  source                        = "../../../terraform/bundles/gcp/master-nodes"
  machine_type                  = var.machine_type
  name_prefix                   = "${var.name_prefix}-${random_id.name_suffix.hex}"
  nodes                         = var.nodes
  provision_mode                = var.provision_mode
  kloudlite_params              = var.kloudlite_params
  save_ssh_key_to_path          = var.save_ssh_key_to_path
  cloudflare                    = var.cloudflare
  public_dns_host               = var.public_dns_host
  save_kubeconfig_to_path       = var.save_kubeconfig_to_path
  labels                        = var.labels
  label_cloudprovider_region    = var.gcp_region
  network                       = var.network
  service_account               = var.service_account
  machine_state                 = var.machine_state
  k3s_service_cidr              = var.k3s_service_cidr
  k3s_download_url              = var.k3s_download_url
  kloudlite_runner_download_url = var.kloudlite_runner_download_url
}

