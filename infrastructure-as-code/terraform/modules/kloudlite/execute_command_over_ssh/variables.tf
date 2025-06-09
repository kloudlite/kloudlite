variable "ssh_params" {
  description = "SSH parameters for the VM"
  type        = object({
    public_ip   = string
    username    = string
    private_key = string
  })
}

variable "timeout_in_minutes" {
  description = "timeout for command"
  type        = number
  default     = 1
}

variable "pre_command" {
  description = "pre command, it's output is not captured"
  type        = string
  default     = ""
}

variable "command" {
  description = "command to be executed, it's stdout is returned as result"
  type        = string
}