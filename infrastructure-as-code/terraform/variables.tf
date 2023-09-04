variable "aws_access_key" {
  default = ""
}

variable "aws_secret_key" {
  default = ""
}

variable "region" {
  default = "ap-south-1"
}

variable "ssh_private_key" {
  default = ""
}

variable "ssh_public_key" {
  default = ""
}

variable "master_nodes_config" {
  type = object(
    {
      name               = string
      count              = number
      instance_type      = string
      ami                = optional(string)
      availability_zones = list(string)
    }
  )
  default = {
    # don't change anything, except the count, and that too only increase it
    # anything else, will cause embedded etcd data loss
    name               = "kloudlite-production-k8s-master"
    count              = 3
    instance_type      = "c6a.large"
    ami                = "ami-094b48639b9ef3b48"
    availability_zones = ["ap-south-1a", "ap-south-1b", "ap-south-1c"]
  }
}

variable "worker_nodes_config" {
  type = object(
    {
      name               = string
      count              = number
      instance_type      = string
      ami                = optional(string)
      availability_zones = list(string)
    }
  )

  default = {
    name               = "kloudlite-production-k8s-worker"
    count              = 1
    instance_type      = "c6a.large"
    ami                = "ami-094b48639b9ef3b48"
    availability_zones = ["ap-south-1a", "ap-south-1b", "ap-south-1c"]
  }
}

variable "cloudflare_api_token" {
  type    = string
  default = ""
}

variable "cloudflare_zone_id" {
  type    = string
  default = "67f645257a633bd1eb1091facfafba04" // kloudlite.io domain on cloudflare
}

variable "cloudflare_domain" {
  type    = string
  default = "test-prod.kloudlite.io"
}
