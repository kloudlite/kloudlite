module "aws-vpc" {
  source         = "../../../terraform/modules/aws/vpc"
  public_subnets = var.public_subnets
  tags           = var.tags
  vpc_cidr       = var.vpc_cidr
  vpc_name       = var.vpc_name
}