terraform {
  cloud {
    organization = "Kloudlite"

    workspaces {
      name = "kloudlite-iac-production"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }

    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "4.13.0"
    }

    sshcommand = {
      source  = "invidian/sshcommand"
      version = "0.2.0"
    }

    ssh = {
      source  = "loafoe/ssh"
      version = "2.6.0"
    }
  }

  required_version = ">= 1.2.0"
}

provider "aws" {
  region     = var.region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

