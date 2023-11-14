variable "tracker_id" {
  description = "reference_id that should be included in names for each of the created resources"
  type        = string
}

variable "spot_fleet_tagging_role_name" {
  description = "The name of the role that will be used to tag spot fleet instances, we will use it to construct role ARN"
  type        = string
}


variable "node_name" {
  description = "spot fleet node name"
  type        = string
}

variable "availability_zone" {
  description = "availability zone in which to create the node"
  type        = string
}

variable "ami" {
  description = "aws ami"
  type        = string
}

variable "ssh_key_name" {
  description = "ssh_key_name to be used when creating instances. It is the output of aws_key_pair.<var-name>.key_name"
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
  description = "user data base64"
  type        = string
}

variable "cpu_node" {
  description = "specs for cpu node"
  type        = object({
    vcpu = object({
      min = number
      max = number
    })
    memory_per_vcpu = object({
      min = number
      max = number
    })
  })
  default = null
}

variable "gpu_node" {
  description = "specs for gpu node"
  type        = object({
    instance_types = list(string)
  })
  default = null
}

variable "last_recreated_at" {
  description = "timestamp when this resource was last recreated, whenever this value changes instance is recreated"
  type        = number
  default     = 0
}