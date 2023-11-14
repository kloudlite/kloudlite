variable "tracker_id" {
  description = "reference_id that should be included in names for each of the created resources"
  type        = string
}

variable "availability_zone" {
  description = "availability zone for nodepool"
  type        = string
  default     = null
}

variable "ami" {
  description = "aws ami to be used for all nodes created in this nodepool"
  type        = string
}

variable "ssh_key_name" {
  description = "ssh_key_name to be used when creating instances. It is the output of aws_key_pair.<var-name>.key_name"
  type        = string
}

variable "instance_type" {
  description = "aws instance type for this nodepool"
  type        = string
}

variable "security_groups" {
  description = "security groups for all nodes in this nodepool"
  type        = list(string)
}

variable "nvidia_gpu_enabled" {
  description = "is this nodepool nvidia gpu enabled"
  type        = string
}

variable "root_volume_size" {
  description = "root volume size for each of the nodes in this nodepool"
  type        = number
}

variable "root_volume_type" {
  description = "root volume type for each of the nodes in this nodepool"
  type        = string
}

variable "iam_instance_profile" {
  description = "iam instance profile for all nodes in this nodepool"
  type        = string
  default     = null
}

variable "nodes" {
  description = "map of nodes to be created in this nodepool"
  type        = map(object({
    user_data_base64  = optional(string)
    last_recreated_at = optional(number)
  }))
}
