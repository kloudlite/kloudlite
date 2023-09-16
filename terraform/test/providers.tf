resource "local_file" "kubeconfig" {
  #  content = base64decode(module.k3s-primary-master.kubeconfig_with_public_ip)
  content  = base64decode(module.k3s-primary-master.kubeconfig_with_public_ip)
  filename = "/tmp/kubeconfig"
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

provider "helm" {
  kubernetes {
    config_path = local_file.kubeconfig.filename
  }
}
