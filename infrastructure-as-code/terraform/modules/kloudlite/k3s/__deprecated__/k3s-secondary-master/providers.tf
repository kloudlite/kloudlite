terraform {
  required_version = ">= 1.2.0"
  required_providers {
    ssh = {
      source  = "loafoe/ssh"
      version = "2.6.0"
    }
  }
}
