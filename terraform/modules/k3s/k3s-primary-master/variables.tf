variable "node_name" {
  description = "The name of the primary node"
  type        = string
}

variable "public_ip" {
  description = "The IP address of the primary master node"
  type        = string
}

variable "k3s_master_nodes_public_ips" {
  description = "The public IP address of the all the master nodes"
  type        = list(string)
}

variable "public_dns_hostname" {
  description = "The domain name to use for the cluster, e.g. cluster.example.com. It is used for the TLS certificate for etcd the Kubernetes API Server"
  type        = string
}

variable "ssh_params" {
  description = "The SSH parameters to use when connecting to the primary master node"
  type        = object({
    user        = string
    private_key = string
  })
}

variable "node_labels" {
  description = "Labels to be added to the nodes"
  type        = map(string)
  default     = {}
}

variable "node_taints" {
  description = "Taints to be added to the nodes"
  type        = map(object({
    value  = string
    effect = string
  }))
  default = {}
}

variable "backup_to_s3" {
  description = "configuration to backup k3s etcd to s3"
  type        = object({
    enabled = bool

    aws_access_key = optional(string, "")
    aws_secret_key = optional(string, "")

    bucket_name   = optional(string, "")
    bucket_region = optional(string, "")
    bucket_folder = optional(string, "")

    cron_schedule = optional(string, "")
  })

  validation {
    error_message = "when backup_to_s3 is enabled, all the following variables must be set: aws_access_key, aws_secret_key, bucket_name, bucket_region, bucket_folder and cron_schedule"
    condition     = var.backup_to_s3.enabled == false || alltrue([
      var.backup_to_s3.aws_access_key != "",
      var.backup_to_s3.aws_secret_key != "",
      var.backup_to_s3.bucket_name != "",
      var.backup_to_s3.bucket_region != "",
      var.backup_to_s3.bucket_folder != "",
      var.backup_to_s3.cron_schedule != "",
    ])
  }
}

variable "restore_from_latest_s3_snapshot" {
  type    = bool
  default = false
}