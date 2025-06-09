variable "vpc_name" {
  description = "VPC name"
  type        = string
}

variable "vpc_cidr" {
  description = "vpc CIDR"
  type        = string
}

variable "public_subnets" {
  description = "list of public subnets"
  type        = list(object({
    availability_zone = string
    cidr              = string
  }))
}

variable "tags" {
  description = "tags to be attached to resource"
  type        = map(string)
}

