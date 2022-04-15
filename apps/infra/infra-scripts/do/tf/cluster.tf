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

resource "digitalocean_droplet" "master-nodes"  {
  count    = var.master-nodes-count
  image    = var.do-image-id
  name     = "${var.cluster-id}-master-node-${count.index}"
  region = var.region
  size     = var.size
  ssh_keys = var.ssh_keys

  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
    wg_ip  = "10.13.13.${count.index + 2}"
  })
}

resource "digitalocean_droplet" "agent-nodes"  {
  count    = var.agent-nodes-count
  image    = var.do-image-id
  name     = "${var.cluster-id}-agent-node-${count.index}"
  region = var.region
  size     = var.size
  ssh_keys = var.ssh_keys
  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
    wg_ip  = "10.13.13.${count.index + 5}"
  })
}


# resource "digitalocean_floating_ip" "floating-ips" {
#   count = var.master-nodes-count
#   region = var.region
# }

# resource "digitalocean_floating_ip_assignment" "ip-assignments" {
#   count = var.master-nodes-count
#   depends_on = [
#     digitalocean_droplet.master-nodes,
#     digitalocean_floating_ip.floating-ips
#   ]
#   droplet_id = digitalocean_droplet.master-nodes[count.index].id
#   ip_address = digitalocean_floating_ip.floating-ips[count.index].ip_address
# }

# output "master-floating-ips" {
#   value = join(",", digitalocean_floating_ip.floating-ips.*.ip_address)
# }
# output master-wg-ips {
#   count = var.master-nodes-count
#   value = join(",", "10.13.13.${count.index + 2}")
# }

output "master-nodes-count" {
  value = var.master-nodes-count
}

output "agent-nodes-count" {
  value = var.agent-nodes-count
}

output "master-ips" {
  value = join(",", digitalocean_droplet.master-nodes.*.ipv4_address)
}

output "agent-ips" {
  value = join(",", digitalocean_droplet.agent-nodes.*.ipv4_address)
}

output "master-internal-ip" {
  value = digitalocean_droplet.master-nodes[0].ipv4_address
}
