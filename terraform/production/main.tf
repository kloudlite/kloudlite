variable "cloudflare_api_token" {
  type = string
}

variable "cloudflare_domain" {
  type    = string
  default = "test-prod.kloudlite.io"
}

variable "cloudflare_zone_id" {
  default = "67f645257a633bd1eb1091facfafba04"
}

variable "aws_access_key" {
  type = string
}

variable "aws_secret_key" {
  type = string
}

module "k3s-HA-on-ec2" {
  source = "../modules/k3s-HA-on-ec2"

  master_nodes_config = {
    name               = "kloudlite-production-k8s-master"
    count              = 3
    instance_type      = "c6a.large"
    ami                = "ami-094b48639b9ef3b48"
    availability_zones = ["ap-south-1a", "ap-south-1b", "ap-south-1c"]
  }

  worker_nodes_config = {
    name               = "kloudlite-production-k8s-worker",
    count              = 1,
    instance_type      = "c6a.large"
    ami                = "ami-094b48639b9ef3b48"
    availability_zones = ["ap-south-1a", "ap-south-1b", "ap-south-1c"]
  }

  domain = var.cloudflare_domain

  aws_region = "ap-south-1"

  aws_access_key = var.aws_access_key
  aws_secret_key = var.aws_secret_key
}

output "kubeconfig" {
  value = module.k3s-HA-on-ec2.kubeconfig
}

module "cloudflare-dns" {
  source = "../modules/cloudflare-dns"

  cloudflare_api_token = var.cloudflare_api_token
  cloudflare_domain    = var.cloudflare_domain
  cloudflare_zone_id   = var.cloudflare_zone_id

  public_ips = module.k3s-HA-on-ec2.k8s_masters_public_ips
}
