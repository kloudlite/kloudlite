terraform {
  required_version = ">= 1.2.0"

  backend "s3" {
    bucket = "sample"
    key    = "sdfaf"
    region = "ap-south-1"
  }

  required_providers {
    ssh = {
      source  = "loafoe/ssh"
      version = "2.6.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "4.13.0"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "1.14.0"
    }
  }
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

data "template_file" "kubeconfig" {
  template = base64decode(module.aws-k3s-HA.kubeconfig_with_master_public_ip)
}

#resource "local_file" "kubeconfig" {
#  #  content = base64decode(module.k3s-primary-master.kubeconfig_with_public_ip)
#  content  = base64decode(module.aws-k3s-HA.kubeconfig_with_master_public_ip)
#  filename = "/tmp/kubeconfig"
#}

provider "ssh" {
  debug_log = "/tmp/kloudlite-iac-ssh.log"
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

provider "helm" {
  kubernetes {
    config_path = data.template_file.kubeconfig.filename
  }
}

provider "kubectl" {
  #  config_path = local_file.kubeconfig.filename
  config_path = data.template_file.kubeconfig.filename
}
