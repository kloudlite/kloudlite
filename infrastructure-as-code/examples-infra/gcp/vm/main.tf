module "vm" {
  source                      = "../../../terraform/bundles/gcp/vm"
  name_prefix                 = var.name_prefix
  vm_name                     = var.vm_name
  provision_mode              = var.provision_mode
  availability_zone           = var.availability_zone
  network                     = var.network
  service_account             = var.service_account
  machine_type                = var.machine_type
  bootvolume_type             = var.bootvolume_type
  bootvolume_size             = var.bootvolume_size
  labels                      = var.labels
  allow_incoming_http_traffic = var.allow_incoming_http_traffic
  allow_ssh                   = var.allow_ssh
  machine_state               = var.machine_state
  startup_script              = var.startup_script
}
