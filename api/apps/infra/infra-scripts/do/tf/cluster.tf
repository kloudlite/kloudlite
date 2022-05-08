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

resource "digitalocean_droplet" "master-node-instances"  {
  for_each = var.master-nodes
  image    = var.do-image-id
  name     = var.master-node-data[each.value].name
  region = var.region
  size     = var.master-node-size
  ssh_keys = var.ssh_keys
  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
    wg_ip  = var.master-node-data[each.value].ip
  })
}

resource "digitalocean_droplet" "agent-node-instances"  {
  for_each    = var.agent-nodes
  image    = var.do-image-id
  name     = var.agent-node-data[each.value].name
  region = var.region
  size     = var.agent-node-size
  ssh_keys = var.ssh_keys
  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
    wg_ip  = var.agent-node-data[each.value].ip
  })
}

output "master-nodes-count" {
  value = length(var.master-nodes)
}

output "agent-nodes-count" {
  value = length(var.agent-nodes)
}


output "master-ips" {
  value = join(",", [for instance in digitalocean_droplet.master-node-instances : instance.ipv4_address])
}

output "agent-ips" {
  value = join(",", [for name,instance in digitalocean_droplet.agent-node-instances : "${name}:${instance.ipv4_address}"])
}


output "master-internal-ip" {
  value = digitalocean_droplet.master-node-instances["master"].ipv4_address
}
