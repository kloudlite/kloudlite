variable "ami" {
  description = "AWS AMI to use for the nodes"
  type        = string
}

variable "save_ssh_key" {
  type = object({
    enabled = string
    path    = string
  })
  default = null
}

variable "nodes_config" {
  type = map(object({
    instance_type        = string
    az                   = string
    root_volume_size     = number
    root_volume_type     = string // standard, gp2, io1, gp3 etc
    with_elastic_ip      = bool
    security_groups      = list(string)
    iam_instance_profile = optional(string)
  }))
}