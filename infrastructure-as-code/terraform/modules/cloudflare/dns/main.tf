resource "cloudflare_record" "A_records" {
  for_each        = {for idx, value in var.public_ips : idx => value}
  zone_id         = var.cloudflare_zone_id
  name            = var.cloudflare_domain
  value           = each.value
  type            = "A"
  ttl             = var.ttl
  proxied         = var.use_cloudflare_proxy
  allow_overwrite = true
  comment         = "managed by kloudlite's infrastructure-as-code"
}

resource "cloudflare_record" "wildcard_cname" {
  count           = var.set_wildcard_cname ? 1 : 0
  zone_id         = var.cloudflare_zone_id
  name            = "*.${var.cloudflare_domain}"
  value           = var.cloudflare_domain
  type            = "CNAME"
  ttl             = var.ttl
  proxied         = var.use_cloudflare_proxy
  allow_overwrite = true
  comment         = "managed by kloudlite's infrastructure-as-code"
}
