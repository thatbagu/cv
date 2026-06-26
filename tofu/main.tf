# Configure the Google Cloud provider
provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

# Configure the GCS backend
terraform {
  backend "gcs" {
    bucket = var.tf_state_bucket
    prefix = "terraform/state"
  }
}

# Create a GCP Compute Engine instance
resource "google_compute_instance" "website_instance" {
  name         = local.instance_name
  machine_type = var.machine_type

  boot_disk {
    initialize_params {
      image = var.boot_disk_image
    }
  }

  network_interface {
    network = "default"

    access_config {
      nat_ip = google_compute_address.static_ip.address
    }
  }

  metadata = {
    gce-container-declaration = templatefile("${path.module}/container-spec.yaml", {
      docker_image = var.docker_image
    })
  }

  tags = ["http-server", "https-server"]
}

# Create a static IP address
resource "google_compute_address" "static_ip" {
  name = local.static_ip_name
}

# Allow HTTP traffic from GCP load balancer and health check IP ranges only
resource "google_compute_firewall" "allow_http" {
  name    = "${local.firewall_rule_name}-http"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  source_ranges = ["130.211.0.0/22", "35.191.0.0/16"]
  target_tags   = ["http-server"]
}

# Allow HTTPS traffic from GCP load balancer IP ranges only
resource "google_compute_firewall" "allow_https" {
  name    = "${local.firewall_rule_name}-https"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["443"]
  }

  source_ranges = ["130.211.0.0/22", "35.191.0.0/16"]
  target_tags   = ["https-server"]
}

# Create a global IP address for the load balancer
resource "google_compute_global_address" "default" {
  name = "global-website-ip"
}

resource "google_compute_managed_ssl_certificate" "default" {
  name = "website-cert"

  managed {
    domains = [var.domain_name]
  }
}

# Create the managed SSL certificate for the old domain
resource "google_compute_managed_ssl_certificate" "old_domain" {
  name = "old-domain-cert"

  managed {
    domains = ["kosaretsky.co.uk"]
  }
}

# Create the HTTPS load balancer for the primary domain
resource "google_compute_target_https_proxy" "default" {
  name             = "website-target-proxy"
  url_map          = google_compute_url_map.default.id
  ssl_certificates = [google_compute_managed_ssl_certificate.default.id]
}

# Create the HTTPS load balancer for the old domain
resource "google_compute_target_https_proxy" "old_domain" {
  name             = "old-domain-target-proxy"
  url_map          = google_compute_url_map.old_domain.id
  ssl_certificates = [google_compute_managed_ssl_certificate.old_domain.id]
}

# Create a URL map for the primary domain
resource "google_compute_url_map" "default" {
  name            = "website-url-map"
  default_service = google_compute_backend_service.default.id
}

# Create a URL map for the old domain (redirect to new domain)
resource "google_compute_url_map" "old_domain" {
  name = "old-domain-url-map"

  default_url_redirect {
    host_redirect          = var.domain_name
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
    strip_query            = false
  }
}

# Cloud Armor security policy: rate limiting + DDoS protection
resource "google_compute_security_policy" "default" {
  name = "website-security-policy"

  rule {
    action   = "rate_based_ban"
    priority = 1000
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
    rate_limit_options {
      conform_action = "allow"
      exceed_action  = "deny(429)"
      enforce_on_key = "IP"
      rate_limit_threshold {
        count        = 600
        interval_sec = 60
      }
      ban_duration_sec = 60
    }
    description = "Rate limit: 600 req/min per IP, ban 1 min on breach"
  }

  rule {
    action   = "allow"
    priority = 2147483647
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
    description = "Default allow"
  }
}

# Create a backend service
resource "google_compute_backend_service" "default" {
  name        = "website-backend"
  port_name   = "http"
  protocol    = "HTTP"
  timeout_sec = 30

  health_checks = [google_compute_health_check.default.id]

  backend {
    group = google_compute_instance_group.webservers.id
  }
}

# Create a health check
resource "google_compute_health_check" "default" {
  name               = "website-health-check"
  check_interval_sec = 5
  timeout_sec        = 5

  http_health_check {
    port = 80
  }
}

# Create an instance group
resource "google_compute_instance_group" "webservers" {
  name        = "website-instance-group"
  description = "Web servers instance group"
  zone        = var.zone

  instances = [
    google_compute_instance.website_instance.id,
  ]

  named_port {
    name = "http"
    port = 80
  }
}

# Create a global forwarding rule for the primary domain (HTTPS)
resource "google_compute_global_forwarding_rule" "default" {
  name       = "website-forwarding-rule"
  target     = google_compute_target_https_proxy.default.id
  port_range = "443"
  ip_address = google_compute_global_address.default.address
}

# Create a global IP address for the old domain
resource "google_compute_global_address" "old_domain" {
  name = "old-domain-global-ip"
}

# Create a global forwarding rule for the old domain (HTTPS)
resource "google_compute_global_forwarding_rule" "old_domain_https" {
  name       = "old-domain-forwarding-rule-https"
  target     = google_compute_target_https_proxy.old_domain.id
  port_range = "443"
  ip_address = google_compute_global_address.old_domain.address
}

# Create HTTP to HTTPS redirect for the primary domain
resource "google_compute_url_map" "http_redirect" {
  name = "http-redirect-url-map"

  default_url_redirect {
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
    strip_query            = false
  }
}

# Create HTTP to HTTPS redirect for the old domain
resource "google_compute_url_map" "old_domain_http_redirect" {
  name = "old-domain-http-redirect-url-map"

  default_url_redirect {
    host_redirect          = var.domain_name
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
    strip_query            = false
  }
}

# Create HTTP target proxies
resource "google_compute_target_http_proxy" "http_redirect" {
  name    = "http-redirect-target-proxy"
  url_map = google_compute_url_map.http_redirect.id
}

resource "google_compute_target_http_proxy" "old_domain_http_redirect" {
  name    = "old-domain-http-redirect-target-proxy"
  url_map = google_compute_url_map.old_domain_http_redirect.id
}

# Create HTTP forwarding rules
resource "google_compute_global_forwarding_rule" "http_redirect" {
  name       = "http-redirect-forwarding-rule"
  target     = google_compute_target_http_proxy.http_redirect.id
  port_range = "80"
  ip_address = google_compute_global_address.default.address
}

resource "google_compute_global_forwarding_rule" "old_domain_http_redirect" {
  name       = "old-domain-http-redirect-forwarding-rule"
  target     = google_compute_target_http_proxy.old_domain_http_redirect.id
  port_range = "80"
  ip_address = google_compute_global_address.old_domain.address
}
