locals {
  PROVISION_SPOT     = "SPOT"
  PROVISION_STANDARD = "STANDARD"

  STORAGE_TYPE = "pd-ssd"

  tags = concat(var.tags != null ? var.tags : [], [
    "vm-type-${lower(local.PROVISION_STANDARD)}",
  ])

}

locals {
  additional_disks = {
    for name, storage_cfg in (var.additional_disk != null ? var.additional_disk : {}) : name => storage_cfg
  }
}

// Fetch the latest Ubuntu 22.04 LTS image
data "google_compute_image" "ubuntu_2204_image" {
  family  = "ubuntu-2204-lts"
  project = "ubuntu-os-cloud"
}

resource "google_compute_disk" "boot_disk" {
  name  = "${var.name}-boot-disk"
  zone  = var.availability_zone
  // only use an image data source if you're ok with the disk recreating itself with a new image periodically
  image = data.google_compute_image.ubuntu_2204_image.self_link
  size  = var.bootvolume_size
  type  = var.bootvolume_type

  lifecycle {
    ignore_changes = [image]
  }
}

resource "google_compute_disk" "additional_disk" {
  for_each = local.additional_disks
  name     = each.key
  type     = "pd-ssd" // or "pd-standard" for HDD
  zone     = var.availability_zone
  size     = each.value.size
}

resource "google_compute_instance" "standard" {
  count = var.provision_mode == local.PROVISION_STANDARD ? 1 : 0

  name = var.name
  zone = var.availability_zone  // e.g., us-central1-a

  machine_type = var.machine_type

  tags = local.tags

  metadata_startup_script = var.startup_script
  #  metadata_startup_script = <<EMSS
  #  ${var.startup_script}
  #  EMSS

  metadata = {
    block-project-ssh-keys = "TRUE"
    enable-oslogin         = "TRUE"
    "ssh-keys"             = "ubuntu:${var.ssh_key}"
  }

  network_interface {
    network = var.network
    access_config {}
  }

  boot_disk {
    source = google_compute_disk.boot_disk.name
  }

  dynamic "attached_disk" {
    for_each = local.additional_disks
    content {
      source      = google_compute_disk.additional_disk[attached_disk.key].id
      device_name = attached_disk.key
    }
  }

  shielded_instance_config {
    enable_secure_boot = true
  }

  lifecycle {
    ignore_changes = [
      machine_type,
      #      boot_disk[0].initialize_params[0].image,
      #      boot_disk[0].initialize_params[0].size,
      metadata_startup_script
    ]
  }

  dynamic "service_account" {
    for_each = {for k, v in (var.service_account != null? [var.service_account] : []) : k => v}
    content {
      email  = service_account.value.email
      scopes = service_account.value.scopes
    }
  }
}

resource "google_compute_instance" "spot" {
  count = var.provision_mode == local.PROVISION_SPOT ? 1 : 0
  name  = var.name
  zone  = var.availability_zone  // e.g., us-central1-a

  machine_type = var.machine_type

  tags = local.tags

  metadata_startup_script = <<EOMSS
  mkdir -p /tmp/preemption # needed for preemption handlers to work
  ${var.startup_script}
  EOMSS

  metadata = {
    block-project-ssh-keys = "TRUE"
    enable-oslogin         = "TRUE"
    shutdown-script        = "#! /usr/bin/env bash touch /tmp/preemption/about-to-be-deleted"
  }

  boot_disk {
    source = google_compute_disk.boot_disk.name
  }

  network_interface {
    network = var.network
    access_config {}
  }

  shielded_instance_config {
    enable_secure_boot = true
  }

  // Additional settings can be overridden here
  scheduling {
    preemptible                 = true
    automatic_restart           = false
    provisioning_model          = local.PROVISION_SPOT
    instance_termination_action = "DELETE"
  }

  dynamic "service_account" {
    for_each = {for k, v in (var.service_account != null? [var.service_account] : []) : k => v}
    content {
      email  = service_account.value.email
      scopes = service_account.value.scopes
    }
  }
}
