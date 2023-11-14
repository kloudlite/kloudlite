variable "public_dns_host" {
  description = "The domain name to use for the cluster, e.g. cluster.example.com. It is used for the TLS certificate for etcd the Kubernetes API Server"
  type        = string
}

variable "cluster_internal_dns_host" {
  description = "the internal dns host for the cluster, e.g. cluster.local"
  type        = string
}

variable "ssh_params" {
  description = "The SSH parameters to use when connecting to k3s master node"
  type        = object({
    user        = string
    private_key = string
  })
}

variable "node_taints" {
  description = "Taints to be added to the nodes"
  type        = list(object({
    key    = string
    value  = optional(string)
    effect = string
  }))
}

variable "extra_server_args" {
  description = "extra server args to pass to k3s server"
  type        = list(string)
  default     = []
}

variable "master_nodes" {
  description = "map of secondary master nodes to join the cluster"
  type        = map(object({
    role        = string
    public_ip   = string
    node_labels = map(string)
  }))

  validation {
    error_message = "master_nodes must have at least one entry"
    condition     = alltrue([for k, v in var.master_nodes : contains(["primary-master", "secondary-master"], v.role)])
  }
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
  })

  validation {
    error_message = "when backup_to_s3 is enabled, all the following variables must be set: aws_access_key, aws_secret_key, bucket_name, bucket_region, bucket_folder and cron_schedule"
    condition     = var.backup_to_s3.enabled == false || alltrue([
      var.backup_to_s3.aws_access_key != "",
      var.backup_to_s3.aws_secret_key != "",
      var.backup_to_s3.bucket_name != "",
      var.backup_to_s3.bucket_region != "",
      var.backup_to_s3.bucket_folder != "",
    ])
  }
}

variable "restore_from_latest_s3_snapshot" {
  type = bool
}