variable "access_key" {
  default = ""
}

variable "secret_key" {
  default = ""
}

variable "region" {
  default = ""
}

variable "node_name" {
  default = ""
}

variable "instance_type" {
  default = ""
}

variable "private_key_path" {
  default = ""
}

variable "public_key_path" {
  default = ""
}

variable "keys_path" {
  default = ""
}

variable "ami" {
  default = ""
}

variable "master_nodes_config" {
  type = object(
    {
      name               = string
      count              = number
      instance_type      = string
      ami                = optional(string)
      availability_zones = string
    }
  )
}

variable "worker_nodes_config" {
  type = object(
    {
      name               = string
      count              = number
      instance_type      = string
      ami                = optional(string)
      availability_zones = string
    }
  )
}

variable "k3s_token" {
  default = ""
}


variable "cloudflare" {
  type = object(
    {
      api_token = string
      zone_id   = string
      domain    = string
      #      email              = string
      #      api_key            = string
      #      zone_id            = string
      #      domain             = string
      #      subdomain          = string
    }
  )
}
