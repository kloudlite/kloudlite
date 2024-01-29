module "dns-records" {
  source               = "../../terraform/modules/cloudflare/dns"
  A_records            = var.A_records
  CNAME_records        = var.CNAME_records
  TXT_records          = var.TXT_records
  cloudflare_api_token = var.cloudflare_api_token
  cloudflare_zone_id   = var.cloudflare_zone_id
  use_cloudflare_proxy = var.use_cloudflare_proxy
}