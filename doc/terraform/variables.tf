// provider

variable "gcp_credentials_file_path" {
  description = "Location of the Google Cloud Platform credentials to use."
  default     = "~/Credentials/skinny-consensus.gcp-sa.json"
}

variable "gcp_project" {
  description = "ID of the Google Cloud Platform project to use."
  default     = "skinny-consensus"
}

variable "gcp_region" {
  description = "Name of the Google Cloud Platform region to use."
  default      = "us-central1"
}

// instance

variable "ssh_user" {
  default = "danrl"
}

variable "ssh_private_key_path" {
  default = "~/.ssh/id_rsa"
}

// dns

variable "dns_managed_zone" {
  default = "net-cakelie-skinny"
}

variable "dns_domain" {
  default = "skinny.cakelie.net"
}
