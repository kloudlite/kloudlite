variable "k3s_server_dns_hostname" {
  description = "The domain name or ip that points to k3s master nodes"
  type        = string
}

variable "k3s_token" {
  description = "k3s token used to join agent nodes to the k3s cluster"
  type        = string
}

variable "agent_nodes" {
  description = "The list of agent nodes"
  type        = map(object({
    public_ip  = string
    ssh_params = object({
      user        = string
      private_key = string
    })
    node_labels = map(string)
  }))
}

variable "use_cloudflare_nameserver" {
  description = "use cloudflare nameserver: 1.1.1.1"
  default     = false
}
