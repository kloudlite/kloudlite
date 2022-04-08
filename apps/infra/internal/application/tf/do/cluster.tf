terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.19.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "3.1.0"
    }
  }
}

provider "digitalocean" {
  token = var.do-token
}

resource "digitalocean_droplet" "primary-master" {
  image    = var.do-image-id
  name     = "${var.cluster-id}-master-0"
  region   = "blr1"
  size     = "s-4vcpu-8gb"
  ssh_keys = var.ssh_keys
  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
  })
}

resource "digitalocean_droplet" "masters" {
  count    = var.master-nodes-count - 1
  image    = var.do-image-id
  name     = "${var.cluster-id}-master-${count.index}"
  region   = "blr1"
  size     = "s-4vcpu-8gb"
  ssh_keys = var.ssh_keys
  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
  })
}


resource "digitalocean_droplet" "workers" {
  count    = var.agent-nodes-count
  image    = var.do-image-id
  name     = "${var.cluster-id}-agent-${count.index}"
  region   = "blr1"
  size     = "s-4vcpu-8gb"
  ssh_keys = var.ssh_keys
  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
  })
}

output "master-internal-ip" {
  value = digitalocean_droplet.primary-master.internal_ipv4
}

output "master-ips" {
  value = join(",", digitalocean_droplet.masters.*.ipv4_address)
}

output "agent-ips" {
  value = join(",", digitalocean_droplet.workers.*.ipv4_address)
}

output "master-ip" {
  value = digitalocean_droplet.primary-master.ipv4_address
}

