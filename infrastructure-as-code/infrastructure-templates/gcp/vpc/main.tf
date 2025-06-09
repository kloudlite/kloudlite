resource "random_id" "name_suffix" {
  keepers = {
    vpc_name = var.vpc_name
  }

  byte_length = 4
}


module "vpc" {
  source   = "../../../terraform/modules/gcp/vpc"
  vpc_name = "${var.vpc_name}-${random_id.name_suffix.hex}"
}
