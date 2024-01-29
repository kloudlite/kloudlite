locals {
  allow_overwrite   = true
  dns_entry_comment = "managed by kloudlite's infrastructure-as-code. It should not be touched and replaced manually"
}

resource "cloudflare_record" "A_records" {
  for_each        = var.A_records
  zone_id         = var.cloudflare_zone_id
  name            = each.value.value
  value           = each.key
  type            = "A"
  ttl             = each.value.ttl
  proxied         = var.use_cloudflare_proxy
  allow_overwrite = local.allow_overwrite
  comment         = local.dns_entry_comment
}

resource "cloudflare_record" "CNAME_records" {
  for_each        = var.CNAME_records
  zone_id         = var.cloudflare_zone_id
  name            = each.key
  value           = each.value.value
  type            = "CNAME"
  ttl             = each.value.ttl
  proxied         = var.use_cloudflare_proxy
  allow_overwrite = local.allow_overwrite
  comment         = local.dns_entry_comment
}

resource "cloudflare_record" "TXT_records" {
  for_each        = var.TXT_records
  zone_id         = var.cloudflare_zone_id
  name            = each.key
  value           = each.value.value
  type            = "TXT"
  ttl             = each.value.ttl
  proxied         = var.use_cloudflare_proxy
  allow_overwrite = local.allow_overwrite
  comment         = local.dns_entry_comment
}
