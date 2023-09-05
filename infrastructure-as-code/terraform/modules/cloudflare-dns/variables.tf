variable "cloudflare_api_token" {
  type    = string
  default = ""
}

variable "cloudflare_zone_id" {
  type    = string
  default = "" // kloudlite.io domain on cloudflare
}

variable "cloudflare_domain" {
  type    = string
  default = ""
}

variable "public_ips" {
  type = list(string)
  default = []
}
