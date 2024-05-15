module "ssh-rsa-key" {
  source = "../../../modules/common/ssh-rsa-key"
}

resource "null_resource" "save_ssh_key" {
  count = length(var.save_ssh_key_to_path) > 0 ? 1 : 0

  provisioner "local-exec" {
    quiet   = true
    command = "echo '${module.ssh-rsa-key.private_key}' > ${var.save_ssh_key_to_path} && chmod 600 ${var.save_ssh_key_to_path}"
  }
}

module "constants" {
  source = "../../../modules/common/constants"
}

module "kloudlite-k3s-templates" {
  source = "../../../modules/kloudlite/k3s/k3s-templates"
}

locals {
  k3s_masters_tags = ["${var.name_prefix}-k3s-masters"]
}

module "master-nodes-firewall" {
  source = "../../../modules/gcp/firewall"

  name_prefix                 = "${var.name_prefix}-fw"
  for_master_nodes            = true
  allow_incoming_http_traffic = true
  allow_node_ports            = true
  network_name                = var.network
  target_tags                 = local.k3s_masters_tags
  allow_ssh                   = true
}

module "master-nodes" {
  source = "../../../modules/gcp/machine"

  for_each = { for name, cfg in var.nodes : name => cfg }

  machine_type      = var.machine_type
  name              = "${var.name_prefix}-${each.key}"
  network           = var.network
  provision_mode    = var.provision_mode
  ssh_key           = module.ssh-rsa-key.public_key
  availability_zone = each.value.availability_zone

  network_tags = local.k3s_masters_tags
  labels       = var.labels

  startup_script = templatefile(module.kloudlite-k3s-templates.k3s-vm-setup-template-path, {
    # kloudlite_release          = var.kloudlite_params.release
    k3s_download_url              = var.k3s_download_url
    kloudlite_runner_download_url = var.kloudlite_runner_download_url
    kloudlite_config_directory    = module.kloudlite-k3s-templates.kloudlite_config_directory
  })
  bootvolume_size = each.value.bootvolume_size
  bootvolume_type = each.value.bootvolume_type
  service_account = var.service_account

  machine_state = var.machine_state
}

module "kloudlite-k3s-masters" {
  source = "../../kloudlite-k3s-masters"
  backup_to_s3 = {
    enabled = false,
  }
  cloudflare                = var.cloudflare
  cluster_internal_dns_host = var.cluster_internal_dns_host
  enable_nvidia_gpu_support = false
  kloudlite_params          = var.kloudlite_params
  master_nodes = {
    for name, cfg in var.nodes : name => {
      role : cfg.k3s_role,
      public_ip : module.master-nodes[name].public_ip,
      node_labels : merge(cfg.node_labels, {
        (module.constants.node_labels.kloudlite_release) : cfg.kloudlite_release,
        (module.constants.node_labels.provider_name) : "gcp",
        (module.constants.node_labels.provider_region) : var.label_cloudprovider_region,
        (module.constants.node_labels.provider_az) : cfg.availability_zone,
        (module.constants.node_labels.node_has_role) : cfg.k3s_role,
        (module.constants.node_labels.node_is_master) : "true",
        #        (module.constants.node_labels.provider_aws_instance_profile_name) : "",
        (module.constants.node_labels.provider_instance_type) : var.machine_type,
        },
      var.use_as_longhorn_storage_nodes ? { "node.longhorn.io/create-default-disk" = "true" } : {}),
      availability_zone = cfg.availability_zone,
      last_recreated_at = 1,
      kloudlite_release = cfg.kloudlite_release,
    }
  }
  public_dns_host              = var.public_dns_host
  restore_from_latest_snapshot = false
  ssh_private_key              = module.ssh-rsa-key.private_key
  ssh_username                 = "ubuntu"
  taint_master_nodes           = true
  tracker_id                   = var.name_prefix
  save_kubeconfig_to_path      = var.save_kubeconfig_to_path
  cloudprovider_name           = "gcp"
  cloudprovider_region         = ""
  k3s_service_cidr             = var.k3s_service_cidr
}
