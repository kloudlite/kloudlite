locals {
  ssh_username = "ubuntu"
}

output "ssh_username" {
  value = local.ssh_username
}

output "public_ip" {
  value = var.provision_mode == local.PROVISION_STANDARD ? google_compute_instance.standard[0].network_interface[0].access_config[0].nat_ip : ""
}

output "machine_id" {
  value = var.provision_mode == local.PROVISION_STANDARD ? google_compute_instance.standard[0].id : google_compute_instance.spot[0].id
}

