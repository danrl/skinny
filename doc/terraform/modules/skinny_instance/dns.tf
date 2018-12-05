// dns

resource "google_dns_record_set" "default" {
  name = "${var.name}.${var.dns_domain}."
  type = "A"
  ttl  = 300

  managed_zone = "${var.dns_managed_zone}"

  rrdatas = ["${google_compute_instance.default.network_interface.0.access_config.0.nat_ip}"]
}
