variable "aws_access_key" { type = string }
variable "aws_secret_key" { type = string }

variable "aws_region" { type = string }
variable "aws_ami" { type = string }

variable "aws_ami_ssh_username" {
  description = "aws ami ssh username"
  type        = string
}

variable "aws_iam_instance_profile_role" {
  description = "aws iam instance profile role"
  type        = string
  default     = ""
}

variable "k3s_server_dns_hostname" {
  description = "The domain name or ip that points to k3s master nodes"
  type        = string
}

variable "cloudflare" {
  description = "use cloudflare nameserver: 1.1.1.1"
  type        = object({
    enabled   = bool
    api_token = optional(string)
    domain    = optional(string)
    zone_id   = optional(string)
  })

  validation {
    condition = var.cloudflare.enabled || (
    var.cloudflare.api_token != "" && var.cloudflare.domain != "" && var.cloudflare.zone_id != ""
    )
    error_message = "when cloudflare is enabled, api_token, domain and zone_id are required"
  }
}

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
      for k, v in var.ec2_nodes_config : contains(["primary-master", "secondary-master", "agent"], v.role)
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
    condition     = !var.spot_settings.enabled || var.spot_settings.spot_fleet_tagging_role_name != ""
    error_message = "when spot_settings is enabled, spot_fleet_tagging_role_name is required"
  }
}

variable "spot_nodes_config" {
  description = "spot nodes configuration"
  type        = map(object({
    az   = optional(string)
    vcpu = object({
      min = number
      max = number
    })
    memory_per_vcpu = object({
      min = number
      max = number
    })
    root_volume_size = optional(number, 50)
    root_volume_type = optional(string, "gp3")
  }))
}

variable "disable_ssh" {
  description = "disable ssh access to the nodes"
  type        = bool
  default     = true
}

variable "k3s_backup_to_s3" {
  description = "configuration to backup k3s etcd to s3"
  type        = object({
    enabled = bool

    bucket_name   = optional(string)
    bucket_region = optional(string)
    bucket_folder = optional(string)
  })

  validation {
    error_message = "when backup_to_s3 is enabled, all the following variables must be set: bucket_name, bucket_region, bucket_folder"
    condition     = var.k3s_backup_to_s3.enabled == false || alltrue([
      var.k3s_backup_to_s3.bucket_name != "",
      var.k3s_backup_to_s3.bucket_region != "",
      var.k3s_backup_to_s3.bucket_folder != "",
    ])
  }
}

variable "restore_from_latest_s3_snapshot" {
  description = "should we restore cluster from latest snapshot"
  type        = bool
  default     = false
}
