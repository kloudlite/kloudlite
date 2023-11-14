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
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }
}