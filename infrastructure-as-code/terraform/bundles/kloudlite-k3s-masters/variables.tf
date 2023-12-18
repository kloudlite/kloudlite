variable "tracker_id" {
  description = "tracker id, for which this resource is being created"
  type        = string
}

variable "ssh_username" {
  description = "ssh username"
  type        = string
}

variable "ssh_private_key" {
  description = "ssh private key"
  type        = string
}

variable "taint_master_nodes" {
  description = "taint master nodes"
  type        = bool
}

variable "public_dns_host" {
  description = "public dns host"
  type        = string
}

variable "cluster_internal_dns_host" {
  description = "cluster internal dns host"
  type        = string
}

variable "backup_to_s3" {
  type = object({
    enabled = bool

    bucket_name   = optional(string)
    bucket_region = optional(string)
    bucket_folder = optional(string)
  })

  validation {
    error_message = "when backup_to_s3 is enabled, all the following variables must be set: aws_access_key, aws_secret_key, bucket_name, bucket_region, bucket_folder"
    condition     = var.backup_to_s3.enabled == false || alltrue([
      var.backup_to_s3.bucket_name != "",
      var.backup_to_s3.bucket_region != "",
      var.backup_to_s3.bucket_folder != "",
    ])
  }
}

variable "restore_from_latest_snapshot" {
  description = "restore from latest snapshot"
  type        = bool
}

variable "cloudflare" {
  description = "cloudflare related parameters"
  type        = object({
    enabled   = bool
    api_token = optional(string)
    zone_id   = optional(string)
    domain    = optional(string)
  })

  validation {
    error_message = "if enabled, all mandatory Cloudflare bucket details are specified"
    condition     = var.cloudflare == null || (var.cloudflare.enabled == true && alltrue([
      var.cloudflare.api_token != "",
      var.cloudflare.zone_id != "",
      var.cloudflare.domain != "",
    ]))
  }
}

variable "master_nodes" {
  description = "k3s masters configuration"
  type        = map(object({
    role              = string
    public_ip         = string
    node_labels       = map(string)
    availability_zone = string
    last_recreated_at = optional(number)
  }))
}

variable "kloudlite_params" {
  description = "kloudlite related parameters"
  type        = object({
    release            = string
    install_crds       = optional(bool, true)
    install_csi_driver = optional(bool, false)
    install_operators  = optional(bool, false)

    install_agent       = optional(bool, false)
    install_autoscalers = optional(bool, true)
    agent_vars          = optional(object({
      account_name             = string
      cluster_name             = string
      cluster_token            = string
      message_office_grpc_addr = string
    }))
  })

  validation {
    error_message = "description"
    condition     = var.kloudlite_params.install_agent == false || var.kloudlite_params.agent_vars != null
  }
}

variable "extra_server_args" {
  description = "extra server args to pass to k3s server"
  type        = list(string)
  default     = []
}

variable "enable_nvidia_gpu_support" {
  description = "enable nvidia gpu support"
  type        = bool
}

variable "save_kubeconfig_to_path" {
  description = "save kubeconfig to this path"
  type        = string
}


variable "cloudprovider_name" {
  type = string
}

variable "cloudprovider_region" {
  type = string
}