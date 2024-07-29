module "dns-records" {
  source               = "../../terraform/modules/cloudflare/dns"
  cloudflare_api_token = var.cloudflare_api_token
  cloudflare_zone_id   = var.cloudflare_zone_id
  use_cloudflare_proxy = var.use_cloudflare_proxy
  DNS_records          = var.DNS_records
}
