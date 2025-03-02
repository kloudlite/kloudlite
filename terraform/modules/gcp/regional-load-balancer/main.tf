# # Step 1: Create a health check
# resource "google_compute_region_health_check" "http_health_check" {
#   name = "http-health-check"
#
#   http_health_check {
#     port = var.http_port
#   }
# }
#
# # Step 2: Create an instance group
# resource "google_compute_instance_group" "group" {
#   name        = "${var.name}-instance-group"
#   description = "Instance group for servers"
#   zone        = "asia-south1-a"
#
#   instances = var.instances
#
#   named_port {
#     name = "http"
#     port = var.http_port
#   }
#
#   named_port {
#     name = "https"
#     port = var.https_port
#   }
# }
#
# # Step 3: creating default backend service
# resource "google_compute_region_backend_service" "default" {
#   region                = var.gcp_region
#   name                  = "region-service"
#   health_checks         = [google_compute_region_health_check.http_health_check.id]
#   protocol              = "HTTP"
#   load_balancing_scheme = "EXTERNAL"
# }
#
# # URL Map for HTTP Traffic
# resource "google_compute_region_url_map" "url_map" {
#   timeouts {
#     create = "30s"
#   }
#   name            = "http-url-map"
#   region          = var.gcp_region
#   default_service = google_compute_region_backend_service.default.id
# }
#
# # Step 5: Reserve a Regional Static IP for the Load Balancer
# resource "google_compute_address" "lb_ip" {
#   name   = "lb-ip"
#   region = var.gcp_region
# }
#
# # Step 6: Regional Forwarding Rule for HTTP
# resource "google_compute_forwarding_rule" "http_forwarding_rule" {
#   name        = "http-forwarding-rule"
#   region      = var.gcp_region
#   target      = google_compute_target_http_proxy.http_proxy.id
#   port_range  = "80"
#   ip_protocol = "TCP"
#
#   load_balancing_scheme = "EXTERNAL"
# }
#
# # Step 7: Regional Forwarding Rule for HTTPS Passthrough
# resource "google_compute_forwarding_rule" "https_forwarding_rule" {
#   name        = "https-forwarding-rule"
#   region      = var.gcp_region
#   target      = google_compute_region_backend_service.default.id
#   port_range  = "443"
#   ip_protocol = "TCP"
#
#   load_balancing_scheme = "EXTERNAL"
# }
#
# # Step 8: Target HTTP Proxy
# resource "google_compute_target_http_proxy" "http_proxy" {
#   name    = "http-proxy"
#   url_map = google_compute_region_url_map.url_map.id
# }

module "load_balancer" {
  source       = "terraform-google-modules/lb/google"
  version      = "5.0.0"
  region       = var.gcp_region
  name         = "${var.name_prefix}-lb"
  service_port = 80
  target_tags  = var.target_tags
  network      = var.network
}

