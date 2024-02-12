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

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key

  dynamic "assume_role" {
    for_each = {
      for idx, config in [var.aws_assume_role != null ? var.aws_assume_role : { enabled : false }] : idx => config
      if config.enabled == true
    }
    content {
      role_arn     = var.aws_assume_role.role_arn
      session_name = "terraform-session"
      external_id  = var.aws_assume_role.external_id
    }
  }
}

provider "ssh" {
  debug_log = "/tmp/kl-target-cluster-aws-ssh-debug.log"
}

provider "cloudflare" {
  api_token = (var.k3s_masters.cloudflare != null && var.k3s_masters.cloudflare.enabled) ? var.k3s_masters.cloudflare.api_token : null
}
