terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.22.3"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.0.3"
    }
  }
}

provider "digitalocean" {
  token = var.do-token
}

resource "digitalocean_droplet" "byoc-node"  {
  image    = var.do-image-id
  name     = var.nodeId
  region = var.region
  size     = var.size
  ssh_keys = var.ssh_keys
  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
  })

}

output "node-ip" {
  value =  digitalocean_droplet.byoc-node.ipv4_address
}

output "node-name" {
  value = digitalocean_droplet.byoc-node.name
}
