variable "primary_master_public_ip" {
  description = "The public IP address of the primary master node"
  type        = string
}

variable "k3s_token" {
  description = "The token to use to join primary k3s cluster as secondary masters"
  type        = string
}

variable "public_dns_hostname" {
  description = "The domain name to use for the cluster, e.g. cluster.example.com. It is used for the TLS certificate for etcd the Kubernetes API Server"
  type        = string
}

variable "cluster_internal_dns_host" {
  description = "the internal dns host for the cluster, e.g. cluster.local"
  type        = string
}

variable "ssh_params" {
  type = object({
    public_ip   = string
    user        = string
    private_key = string
  })
}

variable "node_name" {
  description = "node name"
  type        = string
}

variable "node_labels" {
  description = "node labels"
  type        = map(string)
}

variable "node_taints" {
  description = "Taints to be added to the nodes"
  type        = list(object({
    key    = string
    value  = optional(string)
    effect = string
  }))
  default = []
}


#variable "secondary_masters" {
#  description = "A map of secondary master nodes to join to the primary master node e.g. <node-name> = {} "
#  type        = map(object({
#    public_ip   = string
#    #    ssh_params  =
#    node_labels = map(string)
#    node_taints = map(object({
#      value  = string
#      effect = string
#    }))
#    k3s_backup_cron_schedule = optional(string)
#    is_nvidia_gpu_node       = optional(bool)
#  }))
#}

#variable "k3s_master_nodes_public_ips" {
#  description = "A list of private IP addresses of the k3s masters"
#  type        = list(string)
#}

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
