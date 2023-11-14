variable "aws_region" { type = string }

variable "tracker_id" {
  description = "tracker id, for which this resource is being created"
  type        = string
}

variable "k3s_join_token" {
  description = "k3s join token, that should be used to join the cluster"
  type        = string
}

variable "k3s_server_public_dns_host" {
  description = "k3s server public dns host, i.e. k3s server public url"
  type        = string
}

variable "ec2_nodepools" {
  type = map(object({
    image_id             = string
    image_ssh_username   = string
    availability_zone    = optional(string)
    instance_type        = string
    nvidia_gpu_enabled   = optional(bool)
    root_volume_size     = string
    root_volume_type     = string
    iam_instance_profile = optional(string)
    node_taints          = optional(list(object({
      key    = string
      value  = optional(string)
      effect = string
    })))
    nodes = map(object({
      last_recreated_at = optional(number)
    }))
  }))
}

variable "spot_nodepools" {
  type = map(object({
    image_id                     = string
    image_ssh_username           = string
    availability_zone            = optional(string)
    root_volume_size             = string
    root_volume_type             = string
    iam_instance_profile         = optional(string)
    spot_fleet_tagging_role_name = string

    cpu_node = optional(object({
      vcpu = object({
        min = number
        max = number
      })
      memory_per_vcpu = object({
        min = number
        max = number
      })
    }))

    gpu_node = optional(object({
      instance_types = list(string)
    }))

    node_taints = optional(list(object({
      key    = string
      value  = optional(string)
      effect = string
    })))

    nodes = map(object({
      last_recreated_at = optional(number)
    }))
  }))

  validation {
    error_message = "a nodepool can be either a cpu_node or a gpu_node, only one of them can be set at once"
    condition     = alltrue([
      for name, config in var.spot_nodepools :
      ((config.cpu_node == null && config.gpu_node != null) || (config.cpu_node != null && config.gpu_node == null))
    ])
  }
}

variable "extra_agent_args" {
  description = "extra agent args to pass to k3s agent"
  type        = list(string)
}

variable "save_ssh_key_to_path" {
  description = "save ssh key to this path"
  type        = string
}
