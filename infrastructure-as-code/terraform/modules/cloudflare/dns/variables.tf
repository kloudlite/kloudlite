variable "cloudflare_api_token" {
  description = "cloudflare api token"
  type        = string
}

variable "cloudflare_zone_id" {
  description = "cloudflare zone id"
  type        = string
}

variable "cloudflare_domain" {
  description = "cloudflare domain"
  type        = string
}

variable "public_ips" {
  description = "list of public ips"
  type        = list(string)
}

variable "set_wildcard_cname" {
  description = "should we set a wildcard cname for provided domain at *.domain"
  type        = bool
}

variable "use_cloudflare_proxy" {
  description = "should we use cloudflare proxy for provided domain"
  type        = bool
  default     = false
}

variable "ttl" {
  description = "ttl for dns records"
  type        = number
  default     = 120
}