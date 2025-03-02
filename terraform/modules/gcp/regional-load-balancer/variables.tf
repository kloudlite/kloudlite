variable "name_prefix" {
  type        = string
  description = "name to be used as prefix"
}

variable "gcp_region" {
  type        = string
  description = "GCP region to create regional application load balancer"
}

variable "target_tags" {
  type        = list(string)
  description = "list of target tags to forward traffic to"
}

variable "network" {
  type        = string
  description = "VPC network name"
}
