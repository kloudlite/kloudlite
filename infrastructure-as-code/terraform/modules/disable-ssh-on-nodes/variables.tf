variable "nodes_config" {
  description = "Nodes configuration for whom to disable ssh"
  type = map(object({
    public_ip  = string
    ssh_params = object({
      user        = string
      private_key = string
    })
    disable_ssh = bool
  }))
}