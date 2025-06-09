module "vpc" {
  source   = "../../../terraform/modules/gcp/vpc"
  vpc_name = var.vpc_name
}

