locals {
  vm_group_tags = ["${var.name_prefix}-${var.vm_name}"]
}

module "ssh-rsa-key" {
  source = "../../../modules/common/ssh-rsa-key"
}

module "vm-firewall" {
  source = "../../../modules/gcp/firewall"

  for_vm_group                = true
  network_name                = var.network
  target_tags                 = local.vm_group_tags
  allow_incoming_http_traffic = var.allow_incoming_http_traffic
  allow_node_ports            = false
  name_prefix                 = "${var.name_prefix}-${var.vm_name}-fw"
  allow_ssh                   = var.allow_ssh
}

module "vm-group-nodes" {
  source = "../../../modules/gcp/machine"

  machine_type      = var.machine_type
  service_account   = var.service_account
  name              = "${var.name_prefix}-${var.vm_name}"
  provision_mode    = var.provision_mode
  ssh_key           = module.ssh-rsa-key.public_key
  availability_zone = var.availability_zone
  network           = var.network

  network_tags = local.vm_group_tags
  labels       = var.labels

  startup_script  = var.startup_script
  bootvolume_type = var.bootvolume_type
  bootvolume_size = var.bootvolume_size
  additional_disk = {}
  machine_state   = var.machine_state
}
