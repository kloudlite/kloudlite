variable "ssh_params" {
  description = "The IP address of the primary master node"
  type        = object({
    public_ip   = string
    username    = string
    private_key = string
  })
}

variable "release_name" {
  description = "helm release name"
  type        = string
}
variable "release_namespace" {
  description = "helm release namespace"
  type        = string
}

variable "node_selector" {
  description = "node selector for ebs controller and daemon sets"
  type        = map(string)
}

variable "kloudlite_release" {
  description = "kloudlite release version"
  type        = string
}
