terraform {
  required_version = ">= 1.2.0"
}

provider "aws" {
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = var.aws_region

  dynamic "assume_role" {
    for_each = {for idx, config in [var.aws_assume_role] : idx => config if config.enabled == true}
    content {
      role_arn     = var.aws_assume_role.role_arn
      session_name = "terraform-session"
      external_id  = var.aws_assume_role.external_id
    }
  }
}