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

resource "digitalocean_droplet" "masters" {
  count    = var.master-nodes-count
  image    = "ubuntu-20-04-x64"
  name     = "${var.cluster-id}-master-${count.index}"
  region   = "blr1"
  size     = "s-1vcpu-2gb"
  ssh_keys = var.ssh_keys
  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
  })
}


resource "digitalocean_droplet" "workers" {
  count    = var.agent-nodes-count
  image    = "ubuntu-20-04-x64"
  name     = "${var.cluster-id}-agent-${count.index}"
  region   = "blr1"
  size     = "s-1vcpu-2gb"
  ssh_keys = var.ssh_keys
  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
  })
}

module "k3s" {
  depends_on = [digitalocean_droplet.masters, digitalocean_droplet.workers]
  source  = "xunleii/k3s/module"
  version = "3.1.0"
  servers = {
  for instance in digitalocean_droplet.masters:
    instance.name => {
      ip = instance.ipv4_address_private
      connection = {
        host = instance.ipv4_address
        user = "root"
        private_key = file("${var.keys-path}/access")
      }
    labels = {"node.kubernetes.io/type" = "master"}
    }
  }
  agents = {
    for instance in digitalocean_droplet.workers:
      instance.name => {
        ip = instance.ipv4_address_private
        connection = {
          host = instance.ipv4_address
          user = "root"
          private_key = file("${var.keys-path}/access")
        }
        labels = {"node.kubernetes.io/type" = "agent"}
      }
    }
}

output "kube_config" {
  value     = module.k3s.kube_config
  sensitive = true
}


