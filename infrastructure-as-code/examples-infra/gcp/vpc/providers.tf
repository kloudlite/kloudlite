terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.19.0"
    }
  }
}

provider "google" {
  # Configuration options
  project     = var.gcp_project_id
  region      = var.gcp_region
  credentials = base64decode(var.gcp_credentials_json)
}