variable "name" {
  type        = string
  description = "instance name"
}

variable "availability_zone" {
  type        = string
  description = "Availability Zone"
}

# variable "ssh_key" {
#   type        = string
#   description = "ssh keys"
# }

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
    enabled = bool
    email   = optional(string)
    scopes  = optional(list(string))
  })

  validation {
    error_message = "when service_account enabled, email and scopes must be set"
    condition = anytrue([
      !var.service_account.enabled,
      var.service_account.email != "" && var.service_account.scopes != null
    ])
  }
}

variable "machine_state" {
  type        = string
  description = "state of machine, whether on or off"
  default     = "on"
  validation {
    error_message = "machine_state must be on or off"
    condition = anytrue([
      var.machine_state == "on",
      var.machine_state == "off",
    ])
  }
}

variable "startup_script" {
  type        = string
  description = "startup script"
}

variable "ssh_key" {
  type        = string
  description = "additional ssh keys to attach with VM"
  default     = ""
}

variable "bootvolume_type" {
  type        = string
  description = "bootvolume type like pd-ssd,pd-balanced etc. Read more here https://cloud.google.com/compute/docs/disks"
}

variable "bootvolume_size" {
  type        = number
  description = "bootvolume size in GBs"
}

variable "bootvolume_autodelete" {
  type        = bool
  description = "auto delete bootvolume on instance deletion"
  default     = false
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
}

variable "network_tags" {
  type        = list(string)
  description = "network_tags on compute instance"
  default     = []
}

variable "labels" {
  type        = map(string)
  description = "labels on compute instance"
  default     = {}
}
