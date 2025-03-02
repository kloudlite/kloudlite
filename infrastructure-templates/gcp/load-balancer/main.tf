provider "google" {
  project = "your-gcp-project-id"
  region  = "us-central1"
}

# Create a health check
resource "google_compute_health_check" "http_health_check" {
  name = "http-health-check"

  http_health_check {
    port = 80
  }
}

# Create an instance group
resource "google_compute_instance_group" "web_servers" {
  name        = "web-servers"
  description = "Instance group for web servers"

  instances = [
    "us-central1-a/your-instance-1",
    "us-central1-b/your-instance-2",
  ]

  named_port {
    name = "http"
    port = 80
  }
}

# Create a backend service
resource "google_compute_backend_service" "web_backend" {
  name        = "web-backend"
  description = "Backend service for web servers"

  backend {
    group = google_compute_instance_group.web_servers.self_link
  }

  health_checks = [google_compute_health_check.http_health_check.self_link]
}

# Create a URL map
resource "google_compute_url_map" "web_url_map" {
  name            = "web-url-map"
  description     = "URL map for web servers"

  default_service = google_compute_backend_service.web_backend.self_link
}

# Create a target HTTP proxy
resource "google_compute_target_http_proxy" "http_proxy" {
  name    = "http-proxy"
  url_map = google_compute_url_map.web_url_map.self_link
}

# Create a target HTTPS proxy
resource "google_compute_target_https_proxy" "https_proxy" {
  name             = "https-proxy"
  url_map          = google_compute_url_map.web_url_map.self_link
  ssl_certificates = ["your-ssl-certificate-self-link"]
}

# Create a global forwarding rule for HTTP
resource "google_compute_global_forwarding_rule" "http_forwarding_rule" {
  name       = "http-forwarding-rule"
  target     = google_compute_target_http_proxy.http_proxy.self_link
  port_range = "80"
}

# Create a global forwarding rule for HTTPS
resource "google_compute_global_forwarding_rule" "https_forwarding_rule" {
  name       = "https-forwarding-rule"
  target     = google_compute_target_https_proxy.https_proxy.self_link
  port_range = "443"
}
