variable "source_ami" {
  description = "source AMI ID"
  type        = string
  default     = "ami-03f4878755434977f"
}

variable "region" {
  description = "primary AMI region"
  type        = string
  default     = "ap-south-1"
}

variable "copy_to_regions" {
  description = "list of aws regions, where we need our AMIs"
  type        = list(string)
  default = [
    "us-east-1",
    "eu-west-1",
  ]
}

variable "dest_ami_name" {
  description = "destination AMI name"
  type        = string
  default     = "kloudlite-aws-ami-cpu"
}
