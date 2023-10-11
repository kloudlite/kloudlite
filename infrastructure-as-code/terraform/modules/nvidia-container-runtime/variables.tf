variable "ssh_params" {
  type = object({
    public_ip   = string
    private_key = string
    user        = string
  })
}

variable "gpu_nodes_selector" {
  description = "gpu node selector if applicable"
  type        = map(string)
  default     = {}
}