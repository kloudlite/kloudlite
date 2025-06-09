variable "ssh_params" {
  description = "The IP address of the primary master node"
  type        = object({
    public_ip   = string
    username    = string
    private_key = string
  })
}

#variable "kubeconfig" {
#  type        = string
#  description = "base64 encoded kubeconfig contents"
#}

variable "storage_classes" {
  description = "Storage classes to be created"
  type        = map(object({
    volume_type = string
    fs_type     = string
  }))

  validation {
    condition     = alltrue([for k, v in var.storage_classes : contains(["gp3"], v.volume_type)])
    error_message = "Allowed values for volume_type are gp3 only"
  }
  validation {
    condition     = alltrue([for k, v in var.storage_classes : contains(["ext4", "xfs"], v.fs_type)])
    error_message = "Allowed values for fs_type are ext4 and xfs only"
  }
}

variable "controller_node_selector" {
  description = "node selector for ebs controller pods"
  type        = map(string)
  default     = {}
}

variable "controller_tolerations" {
  description = "tolerations for ebs controller pods"
  type        = list(object({
    key      = string
    effect   = string
    operator = optional(string)
    value    = optional(string, "")
  }))
  default = []
}

variable "daemonset_node_selector" {
  description = "node selector for ebs daemon sets"
  type        = map(string)
  default     = {}
}