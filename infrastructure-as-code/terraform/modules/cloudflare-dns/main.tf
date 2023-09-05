resource "cloudflare_record" "A_records" {
  for_each = { for idx, value in var.public_ips: idx => value }
  zone_id  = var.cloudflare_zone_id
  name     = var.cloudflare_domain
  value    = each.value
  type     = "A"
  ttl      = 300
  proxied  = false
}
