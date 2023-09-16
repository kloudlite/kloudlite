variable "primary_master_public_ip" {
  description = "The public IP address of the primary master node"
  type        = string
}

variable "k3s_token" {
  description = "The token to use to join primary k3s cluster as secondary masters"
  type        = string
}

variable "public_domain" {
  description = "The domain name to use for the cluster, e.g. cluster.example.com. It is used for the TLS certificate for etcd the Kubernetes API Server"
  type        = string
}

variable "secondary_masters" {
  description = "A map of secondary master nodes to join to the primary master node e.g. <node-name> = {} "
  type        = map(object({
    public_ip  = string
    ssh_params = object({
      user        = string
      private_key = string
    })
    node_labels = map(string)
  }))
}

variable "k3s_master_nodes_public_ips" {
  description = "A list of private IP addresses of the k3s masters"
  type        = list(string)
}

variable "disable_ssh" {
  description = "Disable ssh connection to the k3s secondary masters"
  type        = bool
  default     = true
}