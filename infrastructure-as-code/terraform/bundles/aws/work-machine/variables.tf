variable "trace_id" {
  type        = string
  description = "trace id is the prefix/suffix which is added to each resource getting created"
}

variable "name" {
  type        = string
  description = "work machine name"
}

variable "k3s_server_host" {
  type        = string
  description = "k3s server host"
}

variable "k3s_agent_token" {
  type        = string
  description = "k3s agent token"
}

variable "k3s_version" {
  type        = string
  default     = ""
  description = "k3s version, leave empty for latest"
}

variable "ami" {
  type        = string
  description = "work machine name"
}

variable "instance_type" {
  type        = string
  description = "instance type"
}

variable "availability_zone" {
  type        = string
  description = "instance type"
}

variable "iam_instance_profile" {
  type        = string
  description = "IAM Instance Profile"
  default     = ""
}

variable "vpc_id" {
  type        = string
  description = "VPC ID"
}

variable "root_volume_size" {
  description = "root volume size for each of the nodes in this nodepool"
  type        = number
  default     = 50
}

variable "root_volume_type" {
  description = "root volume type for each of the nodes in this nodepool"
  type        = string
  default     = "gp3"
}

variable "security_group_ids" {
  type        = list(string)
  description = "VPC security group IDs"
}

variable "subnet_id" {
  type        = string
  description = "subnet ID for machine"
}

variable "instance_state" {
  type        = string
  description = "state of instance, must be one of running or stopped"
  default     = "running"

  validation {
    condition     = contains(["running", "stopped"], var.instance_state)
    error_message = "Instance state must be either 'running' or 'stopped'."
  }

}
