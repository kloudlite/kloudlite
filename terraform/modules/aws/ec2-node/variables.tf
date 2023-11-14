variable "tracker_id" {
  description = "reference_id that should be included in names for each of the created resources"
  type        = string
}

variable "node_name" {
  description = "spot fleet node name"
  type        = string
}

variable "ssh_key_name" {
  description = "ssh_key_name to be used when creating instances. It is the output of aws_key_pair.<var-name>.key_name"
  type        = string
}

variable "instance_type" {
  description = "instance type of node to be created"
  type        = string
}

variable "is_nvidia_gpu_node" {
  description = "is this a nvidia gpu enabled node"
  type        = bool
}

variable "availability_zone" {
  description = "availability zone in which to create the node"
  type        = string
}

variable "ami" {
  description = "aws ami"
  type        = string
}

variable "root_volume_size" {
  description = "root volume size for each of the nodes in this nodepool"
  type        = number
}

variable "root_volume_type" {
  description = "root volume type for each of the nodes in this nodepool"
  type        = string
  default     = "gp3"
}

variable "security_groups" {
  description = "security groups for all nodes in this nodepool"
  type        = list(string)
}

variable "iam_instance_profile" {
  description = "iam instance profile for all nodes in this nodepool"
  type        = string
  default     = null
}

variable "user_data_base64" {
  description = "user_data_base64 if applicable"
  type        = string
  default     = ""
}

variable "last_recreated_at" {
  description = "timestamp when this resource was last recreated, whenever this value changes instance is recreated"
  type        = number
  default     = 0
}

#variable "nodes_config" {
#  type = map(object({
#    ami                  = string
#    instance_type        = string
#    az                   = optional(string)
#    root_volume_size     = number
#    root_volume_type     = string // standard, gp2, io1, gp3 etc
#    with_elastic_ip      = bool
#    security_groups      = list(string)
#    iam_instance_profile = optional(string)
#    recreate             = optional(bool)
#  }))
#}