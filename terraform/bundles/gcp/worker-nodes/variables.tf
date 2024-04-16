variable "name_prefix" {
  type        = string
  description = "name prefixes to use for resources"
}

variable "nodepool_name" {
  type        = string
  description = "name prefixes to use for k8s nodes"
}

variable "provision_mode" {
  type = string
}

variable "availability_zone" {
  description = "AZ"
  type        = string
}

variable "network" {
  type        = string
  description = "GCP network to use"
}

variable "bootvolume_type" {
  type        = string
  description = "bootvolume type"
}

variable "bootvolume_size" {
  type        = number
  description = "bootvolume size"
}

variable "nodes" {
  type = map(object({}))
}

variable "node_labels" {
  type    = map(string)
  default = {}
}

variable "machine_type" {
  description = "machine_type"
  type        = string
}

variable "additional_disk" {
  type = map(object({
    size = number
    #    type = string pd-ssd
  }))
  default = null
}

variable "k3s_server_public_dns_host" {
  description = "k3s server public DNS host"
  type        = string
}

variable "k3s_join_token" {
  type = string
}

variable "k3s_extra_agent_args" {
  type = list(string)
}

variable "cluster_internal_dns_host" {
  type    = string
  default = "cluster.local"
}

variable "save_ssh_key_to_path" {
  description = "save ssh key to this path"
  type        = string
  default     = ""
}

variable "kloudlite_release" {
  description = "kloudlite release version"
  type        = string
}

variable "label_cloudprovider_region" {
  type        = string
  description = "cloudprovider region"
  default     = ""
}

variable "labels" {
  type        = map(string)
  description = "map of Key => Value to be tagged along created resources"
  default     = {}
}

variable "allow_incoming_http_traffic" {
  type        = bool
  description = "allow incoming http traffic"
}

variable "service_account" {
  type = object({
    enabled = bool
    email   = optional(string)
    scopes  = optional(list(string))
  })
}

