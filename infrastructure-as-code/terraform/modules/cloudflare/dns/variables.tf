variable "cloudflare_api_token" {
  description = "cloudflare api token"
  type        = string
}

variable "cloudflare_zone_id" {
  description = "cloudflare zone id"
  type        = string
}

variable "DNS_records" {
  description = "DNS Records to add"
  type = list(object({
    record_type = string
    domain      = string
    value       = string
    ttl         = optional(number, 120)
  }))

  validation {
    error_message = "record_type should be a valid DNS record type"
    condition     = alltrue([for item in var.DNS_records : contains(["A", "MX", "CNAME", "TXT", "NS"], item.record_type)])
  }
}

variable "use_cloudflare_proxy" {
  description = "should we use cloudflare proxy for provided domain"
  type        = bool
  default     = false
}
