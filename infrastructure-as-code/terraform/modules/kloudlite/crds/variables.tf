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

variable "force_apply" {
  description = "will apply kloudlite CRDs again, otherwise only applies it on resource creation"
  type        = bool
  default     = false
}