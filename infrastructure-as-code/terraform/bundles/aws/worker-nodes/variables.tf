variable "aws_region" { type = string }

variable "tracker_id" {
  description = "tracker id, for which this resource is being created"
  type        = string
}

variable "nodepool_name" {
  description = "nodepool name"
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

variable "kloudlite_release" {
  description = "kloudlite release, to be installed"
  type        = string
}

variable "k3s_download_url" {
  type        = string
  description = "k3s download URL"
}

variable "kloudlite_runner_download_url" {
  type        = string
  description = "kloudlite runner download URL"
}

variable "vpc_id" {
  description = "vpc id"
  type        = string
}

variable "vpc_subnet_id" {
  description = "vpc subnet id"
  type        = string
}

variable "availability_zone" {
  description = "availability zone"
  type        = string
}

variable "nvidia_gpu_enabled" {
  description = "NVIDIA GPU Enabled?"
  type        = bool
  default     = false
}

variable "node_taints" {
  description = "node taints on nodepool nodes"
  type = list(object({
    key    = string
    value  = optional(string)
    effect = string
  }))
  default = null
}

variable "iam_instance_profile" {
  description = "AWS IAM Instance Profile to use"
  type        = string
  default     = ""
}

variable "ec2_nodepool" {
  description = "EC2 nodepool spec"
  type = object({
    instance_type = string

    root_volume_size = number
    root_volume_type = string

    nodes = map(object({
      last_recreated_at = optional(number)
    }))
  })
  default = null
}

variable "spot_nodepool" {
  description = "SPOT nodepool spec"
  type = object({
    root_volume_size = number
    root_volume_type = string

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

    nodes = map(object({
      last_recreated_at = optional(number)
    }))
  })

  default = null
}

variable "extra_agent_args" {
  description = "extra agent args to pass to k3s agent"
  type        = list(string)
}

variable "save_ssh_key_to_path" {
  description = "save ssh key to this path"
  type        = string
}

variable "tags" {
  description = "a map of key values , that will be attached to cloud provider resources, for easier referencing"
  type        = map(string)
  default     = {}
}
