resource "cloudflare_record" "A_records" {
  for_each        = {for idx, value in var.public_ips : idx => value}
  zone_id         = var.cloudflare_zone_id
  name            = var.cloudflare_domain
  value           = each.value
  type            = "A"
  ttl             = 120
  proxied         = false
  allow_overwrite = true
  comment         = "managed by kloudlite's infrastructure-as-code"
}

resource "cloudflare_record" "wildcard_cname" {
  zone_id         = var.cloudflare_zone_id
  name            = "*.${var.cloudflare_domain}"
  value           = var.cloudflare_domain
  type            = "CNAME"
  ttl             = 120
  proxied         = false
  allow_overwrite = true
  comment         = "managed by kloudlite's infrastructure-as-code"
}
