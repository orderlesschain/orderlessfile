terraform {
  required_providers {
    opennebula = {
      source = "OpenNebula/opennebula"
      version = "0.5.0"
    }
  }
}

variable "opennebula_username" {}
variable "opennebula_password" {}


provider "opennebula" {
  endpoint      = " "
  flow_endpoint = " "
  username      = var.opennebula_username
  password      = var.opennebula_password
}
