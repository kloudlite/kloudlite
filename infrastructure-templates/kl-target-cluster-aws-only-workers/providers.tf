terraform {
  required_version = ">= 1.2.0"
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key == "" ? null : var.aws_access_key
  secret_key = var.aws_secret_key == "" ? null : var.aws_secret_key

  dynamic "assume_role" {
    for_each = {
      for idx, config in [var.aws_assume_role != null ? var.aws_assume_role : { enabled : false }] : idx => config
      if config.enabled == true
    }
    content {
      role_arn     = var.aws_assume_role.role_arn
      session_name = "terraform-session"
      external_id  = var.aws_assume_role.external_id
    }
  }
}
