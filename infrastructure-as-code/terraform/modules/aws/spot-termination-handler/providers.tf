terraform {
  required_version = ">= 1.2.0"

  required_providers {
    kubectl = {
      source = "gavinbunney/kubectl"
      version = "1.14.0"
    }
  }
}