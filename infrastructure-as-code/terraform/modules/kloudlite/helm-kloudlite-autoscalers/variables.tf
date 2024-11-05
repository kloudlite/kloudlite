variable "ssh_params" {
  description = "SSH parameters for the VM"
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

variable "kloudlite_release" {
  description = "Kloudlite release to deploy"
  type        = string
}