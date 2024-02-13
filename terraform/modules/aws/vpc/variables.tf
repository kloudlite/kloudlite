variable "tracker_id" {
  type = string
}

variable "vpc_name" {
  type        = string
  description = "VPC name"
}

#variable "vpc_cidr" {
#  type        = string
#  description = "VPC CIDR IP range"
#}

variable "tags" {
  type        = map(string)
  description = "reference tags"
}

#variable "public_subnets" {
#  type = list(object({
#    availability_zone = string
#    cidr              = string
#  }))
#}
