variable "ssh_params" {
  type = object({
    host        = string
    user        = string
    private_key = string
  })
}

variable "release_name" {
  type    = string
  default = "kloudlite-agent"
}

variable "release_namespace" {
  description = "namespace where to install kloudlite-agent"
  type        = string
}

variable "kloudlite_release" {
  description = "kloudlite release version number"
  type        = string
}

variable "args" {
  type = object({
    message_office_grpc_addr = string
    cluster_token            = string

    cluster_name = string
    account_name = string
  })
}