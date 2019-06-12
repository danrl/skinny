provider "google" {
  credentials = "${file("${var.gcp_credentials_file_path}")}"
  project     = "${var.gcp_project}"
  region      = "${var.gcp_region}"

  version     = "~> 2.5"
}
