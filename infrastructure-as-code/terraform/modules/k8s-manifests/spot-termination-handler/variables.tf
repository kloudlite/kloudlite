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