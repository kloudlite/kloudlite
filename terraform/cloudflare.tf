resource "cloudflare_record" "k8s_first_master_A_record" {
  zone_id = var.cloudflare.zone_id
  name    = "test-prod"
  value   = aws_instance.k8s_first_master.public_ip
  type    = "A"
  ttl     = 300
  proxied = false
}

resource "cloudflare_record" "k8s_masters_A_records" {
  for_each = { for idx, value in aws_instance.k8s_masters : idx => value.public_ip }
  zone_id  = var.cloudflare.zone_id
  name     = var.cloudflare.domain
  value    = each.value
  type     = "A"
  ttl      = 300
  proxied  = false
}
