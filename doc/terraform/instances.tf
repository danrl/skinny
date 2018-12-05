module "skinny_instance_oregon" {
  source                = "modules/skinny_instance"
  name                  = "oregon"
  zone                  = "us-west1-a"

  ssh_user              = "${var.ssh_user}"
  ssh_private_key_path  = "${var.ssh_private_key_path}"
  dns_managed_zone      = "${var.dns_managed_zone}"
  dns_domain            = "${var.dns_domain}"
}

module "skinny_instance_spaulo" {
  source  = "modules/skinny_instance"
  name = "spaulo"
  zone = "southamerica-east1-b"

  ssh_user              = "${var.ssh_user}"
  ssh_private_key_path  = "${var.ssh_private_key_path}"
  dns_managed_zone      = "${var.dns_managed_zone}"
  dns_domain            = "${var.dns_domain}"
}

module "skinny_instance_london" {
  source  = "modules/skinny_instance"
  name = "london"
  zone = "europe-west2-c"

  ssh_user              = "${var.ssh_user}"
  ssh_private_key_path  = "${var.ssh_private_key_path}"
  dns_managed_zone      = "${var.dns_managed_zone}"
  dns_domain            = "${var.dns_domain}"
}

module "skinny_instance_taiwan" {
  source  = "modules/skinny_instance"
  name = "taiwan"
  zone = "asia-east1-b"

  ssh_user              = "${var.ssh_user}"
  ssh_private_key_path  = "${var.ssh_private_key_path}"
  dns_managed_zone      = "${var.dns_managed_zone}"
  dns_domain            = "${var.dns_domain}"
}

module "skinny_instance_sydney" {
  source  = "modules/skinny_instance"
  name = "sydney"
  zone = "australia-southeast1-b"

  ssh_user              = "${var.ssh_user}"
  ssh_private_key_path  = "${var.ssh_private_key_path}"
  dns_managed_zone      = "${var.dns_managed_zone}"
  dns_domain            = "${var.dns_domain}"
}
