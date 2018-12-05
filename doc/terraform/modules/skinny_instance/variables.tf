// instance

variable "name" {
  description = "Default name for the Skinny instance."
  default = "skinny-instance"
}
variable "zone" {
  description = "Default zone in which the Skinny instance is started."
  default = "us-central1-a"
}
variable "ssh_user" {
  description = "Username for an SSH connection to the Skinny instance. Most likely your username on Google Cloud Platform."
  default = "ubuntu"
}
variable "ssh_private_key_path" {
  description = "Path to the SSH private key file used for connecting to the Skinny instance."
}

// firewall

variable "port" {
  description = "The port on which the Skinny instance will listen on."
  default = "9000"
}

// dns

variable "dns_managed_zone" {
  description = "Name of the CloudDNS managed zone to which the DNS entries for the Skinny instance shall be added."
}
variable "dns_domain" {
  description = "The domain part of the Skinny instance's fully qualified domain name."
}
