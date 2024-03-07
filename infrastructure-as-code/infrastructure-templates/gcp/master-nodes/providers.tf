terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.19.0"
    }

    ssh = {
      source  = "loafoe/ssh"
      version = "2.6.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "4.13.0"
    }
  }
}

provider "google" {
  # Configuration options
  project     = var.gcp_project_id
  region      = var.gcp_region
  credentials = var.gcp_credentials_json
}

provider "ssh" {
  debug_log = "/tmp/kl-masters-on-gcp.log"
}

provider "cloudflare" {
  api_token = (var.cloudflare != null && var.cloudflare.enabled) ? var.cloudflare.api_token : null
}
