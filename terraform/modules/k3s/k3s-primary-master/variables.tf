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

variable "use_cloudflare_nameserver" {
  description = "use cloudflare nameserver: 1.1.1.1"
  default     = false
}

#variable "cloud_provider_name" {
#  description = "cloud provider name, to be used in node-labels"
#  type        = string
#  default     = ""
#}
#
#variable "cloud_provider_region" {
#  description = "cloud provider region, to be used in node-labels"
#  type        = string
#  default     = ""
#}
#
#variable "cloud_provider_az" {
#  description = "cloud provider az, to be used in node-labels"
#  type        = string
#  default     = ""
#}