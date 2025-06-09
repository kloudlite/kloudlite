variable "gcp_project_id" {
  type        = string
  description = "GCP Project ID"
}

variable "gcp_region" {
  type        = string
  description = "GCP Region"
}

variable "gcp_credentials_json" {
  type        = string
  description = "Credentials JSON"
}


## for nodepools

variable "nodepools" {
  type = map(object({
    provision_mode       = string
    machine_type         = string
    availability_zone    = string
    bootvolume_type      = string
    bootvolume_size      = number
    nodes                = map(object({}))
    node_labels          = map(string)
    k3s_extra_agent_args = list(string)
    additional_disk      = map(object({
      size = number
    }))
  }))
}
