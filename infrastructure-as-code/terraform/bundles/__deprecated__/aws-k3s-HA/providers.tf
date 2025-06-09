terraform {
  required_version = ">= 1.2.0"

  required_providers {
    ssh = {
      source  = "loafoe/ssh"
      version = "2.6.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "4.13.0"
    }
  }
}

resource "local_file" "kubeconfig" {
  filename = "/tmp/kubeconfig"
  content  = base64decode(module.k3s-primary-master.kubeconfig_with_public_host)
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

provider "ssh" {
  debug_log = "/tmp/kloudlite-iac-ssh.log"
}

provider "cloudflare" {
  api_token = var.cloudflare.enabled ? var.cloudflare.api_token : null
}