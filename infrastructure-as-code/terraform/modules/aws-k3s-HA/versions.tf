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
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "1.14.0"
    }
  }
}

#resource "template_file" "kubeconfig" {
#  template = base64encode(module.k3s-primary-master.kubeconfig_with_public_ip)
#}

#data "template_file" "kubeconfig" {
#  template = base64encode(module.k3s-primary-master.kubeconfig_with_public_ip)
#}

resource "null_resource" "kconfig" {
  triggers = {
    always_run = timestamp()
  }

  provisioner "local-exec" {
    quiet   = true
    command = "echo '${base64decode(module.k3s-primary-master.kubeconfig_with_public_ip)}' > /tmp/kubeconfig && chmod 600 /tmp/kubeconfig"
  }
}

resource "local_file" "kubeconfig" {
  filename   = "/tmp/kubeconfig"
  source     = "/tmp/kubeconfig"
  depends_on = [null_resource.kconfig]
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

provider "helm" {
  kubernetes {
    config_path = local_file.kubeconfig.filename
    #    config_path = data.template_file.kubeconfig.filename
  }
}

provider "ssh" {
  debug_log = "/tmp/kloudlite-iac-ssh.log"
}

provider "cloudflare" {
  api_token = var.cloudflare.enabled ? var.cloudflare.api_token : null
}

provider "kubectl" {
  config_path = local_file.kubeconfig.filename
  #  config_path = data.template_file.kubeconfig.filename
}
