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

variable "kloudlite_account_name" {
  description = "Kloudlite account name"
  type        = string
}

variable "kloudlite_cluster_name" {
  description = "Kloudlite cluster name"
  type        = string
}

variable "kloudlite_cluster_token" {
  description = "Kloudlite cluster token"
  type        = string
}

variable "kloudlite_message_office_grpc_addr" {
  description = "Kloudlite message office gRPC address"
  type        = string
}

variable "kloudlite_dns_host" {
  description = "Kloudlite DNS host"
  type        = string
}
