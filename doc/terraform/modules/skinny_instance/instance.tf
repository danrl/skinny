// instance

resource "google_compute_instance" "default" {
  name         = "skinny-${var.name}"
  machine_type = "f1-micro"

  zone = "${var.zone}"

  tags = ["skinny", "skinny-${var.name}"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-1810"
    }
  }

  network_interface {
    network = "default"
    access_config {
      // Ephemeral IP
    }
  }

  service_account {
    scopes = ["userinfo-email", "compute-ro", "storage-ro"]
  }

  provisioner "remote-exec" {
    inline = [
      // includes workaround for a known bug
      // https://github.com/hashicorp/terraform/issues/16656
      "sleep 120",
      "sudo apt-get update",
      "sudo apt-get install -y python", // for ansible
    ]
    connection {
      host        = "${google_compute_instance.default.network_interface.0.access_config.0.nat_ip}"
      type        = "ssh"
      user        = "${var.ssh_user}"
      private_key = "${file(var.ssh_private_key_path)}"
    }
  }

  timeouts {
    create = "10m"
  }
}
