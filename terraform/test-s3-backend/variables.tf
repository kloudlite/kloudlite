variable "aws_access_key" { type = string }
variable "aws_secret_key" { type = string }

variable "aws_region" { type = string }
variable "aws_ami" {
  type    = string
  default = "ami-0f78219c8292792d9"
}

variable "aws_iam_instance_profile_role" {
  description = "aws iam instance profile role"
  type        = string
  default     = ""
}

variable "cloudflare_api_token" { type = string }

variable "cloudflare_domain" { type = string }
variable "cloudflare_zone_id" { type = string }

variable "ec2_nodes_config" {
  description = "ec2 nodes configuration"
  type        = map(object({
    az               = string
    role             = string
    instance_type    = optional(string, "c6a.large")
    root_volume_size = optional(number, 50)
    root_volume_type = optional(string, "gp3")
    with_elastic_ip  = optional(bool, false)
  }))

  validation {
    condition = alltrue([
      for k, v in var.ec2_nodes_config :contains(["primary-master", "secondary-master", "agent"], v.role)
    ])
    error_message = "Invalid node role, must be one of primary, secondary or agent"
  }
}

variable "spot_settings" {
  description = "spot nodes settings"
  type        = object({
    enabled                      = bool
    spot_fleet_tagging_role_name = optional(string)
  })

  validation {
    condition     = var.spot_settings.enabled && var.spot_settings.spot_fleet_tagging_role_name != ""
    error_message = "when spot_settings is enabled, spot_fleet_tagging_role_name is required"
  }
}

variable "spot_nodes_config" {
  description = "spot nodes configuration"
  type        = map(object({
    az               = string
    instance_type    = optional(string, "c6a.large")
    root_volume_size = optional(number, 50)
    root_volume_type = optional(string, "gp3")
    allow_public_ip  = optional(bool, false)
  }))
}

variable "disable_ssh" {
  description = "disable ssh access to the nodes"
  type        = bool
  default     = true
}
