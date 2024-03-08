variable "name" {
  type        = string
  description = "instance name"
}

variable "availability_zone" {
  type        = string
  description = "Availability Zone"
}

variable "ssh_key" {
  type        = string
  description = "ssh keys"
}

variable "provision_mode" {
  type = string
  validation {
    error_message = "must be one of STANDARD or SPOT"
    condition     = contains(["STANDARD", "SPOT"], var.provision_mode)
  }
}

variable "machine_type" {
  type        = string
  description = "machine type"
}

variable "service_account" {
  type = object({
    email  = string
    scopes = list(string)
  })
  default = null
}

variable "startup_script" {
  type        = string
  description = "startup script"
}

variable "bootvolume_type" {
  type        = string
  description = "bootvolume type like pd-ssd,pd-balanced etc. Read more here https://cloud.google.com/compute/docs/disks"
}

variable "bootvolume_size" {
  type        = number
  description = "bootvolume size in GBs"
}

variable "additional_disk" {
  type = map(object({
    size = number
    #    type = string pd-ssd
  }))
  default = null
}

variable "network" {
  type        = string
  description = "network name"
  default     = "default"
}

variable "tags" {
  type        = list(string)
  description = "tags that should be present on resources"
  default     = []
}

