variable "name_prefix" {
  type        = string
  description = "name prefixes to use for resources"
}

variable "vm_name" {
  type        = string
  description = "name prefixes to use for resources"
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

variable "service_account" {
  type = object({
    enabled = bool
    email   = optional(string)
    scopes  = optional(list(string))
  })
}

variable "machine_type" {
  description = "machine_type"
  type        = string
}

variable "bootvolume_type" {
  type        = string
  description = "bootvolume type"
}

variable "bootvolume_size" {
  type        = number
  description = "bootvolume size"
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

variable "allow_ssh" {
  type        = bool
  description = "allow ssh traffic"
}

variable "machine_state" {
  type        = string
  description = "machine state either on or off"
}

variable "startup_script" {
  type        = string
  description = "startup script"
}
