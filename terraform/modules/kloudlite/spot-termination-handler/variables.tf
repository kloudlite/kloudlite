variable "release_name" {
  description = "Name of the release"
  type        = string
}
variable "release_namespace" {
  description = "Namespace of the release"
  type        = string
}

variable "kloudlite_release" {
  description = "Kloudlite release to deploy"
  type        = string
}

variable "spot_nodes_selector" {
  description = "node selector for spot nodes"
  type        = map(string)
}

variable "ssh_params" {
  description = "SSH parameters for the VM"
  type        = object({
    public_ip   = string
    username    = string
    private_key = string
  })
}