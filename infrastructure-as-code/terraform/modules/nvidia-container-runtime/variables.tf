variable "ssh_params" {
  type = object({
    public_ip   = string
    private_key = string
    username    = string
  })
}

variable "gpu_node_selector" {
  description = "gpu node selector if applicable"
  type        = map(string)
}

variable "gpu_node_tolerations" {
  description = "gpu node tolerations if applicable"
  type        = list(object({
    key      = string
    operator = string
    value    = optional(string)
    effect   = string
  }))
}