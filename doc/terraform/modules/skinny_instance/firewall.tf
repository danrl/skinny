// firewall

resource "google_compute_firewall" "default" {
  name    = "skinny-${var.name}"
  network = "default"

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["${var.port}"]
  }

  target_tags = ["skinny-${var.name}"]
  source_ranges = ["0.0.0.0/0"]
}
