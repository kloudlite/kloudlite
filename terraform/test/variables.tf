variable "aws_access_key" { type = string }
variable "aws_secret_key" { type = string }

variable "aws_region" { type = string }
variable "aws_ami" {
  type = string
  default = "ami-0f78219c8292792d9"
}

variable "cloudflare_api_token" { type = string }

variable "cloudflare_domain" { type = string }
variable "cloudflare_zone_id" { type = string }

variable "nodes_config" {
  description = "ec2 nodes configuration"
  type = map(object({
    az               = string
    role             = string
    instance_type    = optional(string, "c6a.large")
    root_volume_size = optional(number, 50)
    root_volume_type = optional(string, "gp3")
    with_elastic_ip   = optional(bool, false)
  }))

  validation {
    condition = alltrue([
      for k, v in var.nodes_config :contains(["primary-master", "secondary-master", "agent"], v.role)
    ])
    error_message = "Invalid node role, must be one of primary, secondary or agent"
  }
}

variable "spot_nodes_config" {
  description = "spot nodes configuration"
  type = map(object({
    az               = string
    instance_type    = optional(string, "c6a.large")
    root_volume_size = optional(number, 50)
    root_volume_type = optional(string, "gp3")
  }))
}