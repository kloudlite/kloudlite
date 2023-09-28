variable "ssh_params" {
  description = "SSH parameters for the VM"
  type        = object({
    public_ip   = string
    username    = string
    private_key = string
  })
}

variable "kloudlite_release" {
  description = "Kloudlite release to deploy"
  type        = string
}