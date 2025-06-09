locals {
  allow_overwrite   = true
  dns_entry_comment = "managed by kloudlite's infrastructure-as-code. It should not be touched and replaced manually"
}

resource "cloudflare_record" "DNS_record" {
  for_each        = { for idx, item in var.DNS_records : idx => item }
  zone_id         = var.cloudflare_zone_id
  name            = each.value.domain
  value           = each.value.value
  type            = each.value.record_type
  ttl             = each.value.ttl
  proxied         = var.use_cloudflare_proxy
  allow_overwrite = local.allow_overwrite
  comment         = local.dns_entry_comment
}

