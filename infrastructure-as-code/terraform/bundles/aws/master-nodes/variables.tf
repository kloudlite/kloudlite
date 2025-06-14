variable "aws_region" {
  description = "aws region"
  type        = string
}

variable "trace_id" {
  description = "trace id, for which this resource is being created"
  type        = string
}

variable "vpc_id" {
  description = "AWs VPC Id"
  type        = string
}

variable "instance_state" {
  description = "Machines Instance State"
  type        = string
}

variable "nodes" {
  description = "list of master nodes"
  type = list(object({
    name              = string
    ami               = string
    instance_type     = string
    availability_zone = string

    security_group_ids = list(string)
    vpc_subnet_id      = string

    root_volume_size = string
    root_volume_type = string
  }))
}

variable "k3s_masters" {
  description = "k3s masters configuration"
  type = object({
    ami           = string
    instance_type = string
    ssh_username  = string

    nvidia_gpu_enabled   = optional(bool)
    root_volume_size     = string
    root_volume_type     = string
    iam_instance_profile = optional(string)
    taint_master_nodes   = bool

    public_dns_host           = string
    cluster_internal_dns_host = string

    backup_to_s3 = object({
      enabled = bool

      # it assumes, access to bucket is managed by IAM instance profile
      endpoint      = optional(string)
      bucket_name   = optional(string)
      bucket_region = optional(string)
      bucket_folder = optional(string)
    })

    restore_from_latest_snapshot = optional(bool, false)

    cloudflare = optional(object({
      enabled   = bool
      api_token = optional(string)
      zone_id   = optional(string)
      domain    = optional(string)
    }))

    nodes = map(object({
      role              = string
      availability_zone = string
      vpc_subnet_id     = string
      last_recreated_at = optional(number)
      kloudlite_release = string
    }))
  })

  validation {
    error_message = "when backup_to_s3 is enabled, all the following variables must be set: endpoint, bucket_name, bucket_region, bucket_folder"
    condition = var.k3s_masters.backup_to_s3.enabled == false || alltrue([
      var.k3s_masters.backup_to_s3.endpoint != "",
      var.k3s_masters.backup_to_s3.bucket_name != "",
      var.k3s_masters.backup_to_s3.bucket_region != "",
      var.k3s_masters.backup_to_s3.bucket_folder != "",
    ])
  }

  validation {
    error_message = "if enabled, all mandatory Cloudflare bucket details are specified"
    condition = var.k3s_masters.cloudflare == null || (var.k3s_masters.cloudflare.enabled == true && alltrue([
      var.k3s_masters.cloudflare.api_token != "",
      var.k3s_masters.cloudflare.zone_id != "",
      var.k3s_masters.cloudflare.domain != "",
    ]))
  }

  validation {
    error_message = "must specify ami, and instance_type"
    condition     = var.k3s_masters.ami != "" && var.k3s_masters.instance_type != ""
  }
}

variable "kloudlite_params" {
  description = "kloudlite related parameters"
  type = object({
    release            = string
    install_crds       = optional(bool, true)
    install_csi_driver = optional(bool, false)
    install_operators  = optional(bool, false)

    install_agent = optional(bool, false)
    agent_vars = optional(object({
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

variable "k3s_service_cidr" {
  type        = string
  description = "k3s service CIDR to use for this cluster, as specified in https://docs.k3s.io/networking/basic-network-options?_highlight=cidr#dual-stack-ipv4--ipv6-networking"
  default     = ""
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

variable "save_ssh_key_to_path" {
  description = "save ssh key to this path"
  type        = string
  default     = ""
}

variable "save_kubeconfig_to_path" {
  description = "save kubeconfig to this path"
  type        = string
  default     = ""
}

variable "tags" {
  description = "a map of key values , that will be attached to cloud provider resources, for easier referencing"
  type        = map(string)
  default     = {}
}

