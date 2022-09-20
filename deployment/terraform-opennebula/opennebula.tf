# Variables
variable "template_id" {}
variable "group_name" {}

variable "node_vm_name" {}
variable "node_counter" {}
variable "node_cpu" {}
variable "node_memory_mb" {}
variable "node_disk_mb" {}

variable "client_vm_name" {}
variable "client_counter" {}
variable "client_cpu" {}
variable "client_memory_mb" {}
variable "client_disk_mb" {}

variable "cli_vm_name" {}
variable "cli_counter" {}
variable "cli_cpu" {}
variable "cli_memory_mb" {}
variable "cli_disk_mb" {}

resource "opennebula_virtual_machine" "node" {
  name        = "${var.node_vm_name}-${count.index}"
  count       = var.node_counter
  cpu         = var.node_cpu
  vcpu        = var.node_cpu
  memory      = var.node_memory_mb
  template_id = var.template_id
  group       = var.group_name

  disk {
    size   = var.node_disk_mb
    target = "vda"
  }
}

resource "opennebula_virtual_machine" "client" {
  name        = "${var.client_vm_name}-${count.index}"
  count       = var.client_counter
  cpu         = var.client_cpu
  vcpu        = var.client_cpu
  memory      = var.client_memory_mb
  template_id = var.template_id
  group       = var.group_name

  disk {
    size   = var.client_disk_mb
    target = "vda"
  }
}

resource "opennebula_virtual_machine" "cli" {
  name        = "${var.cli_vm_name}-${count.index}"
  count       = var.cli_counter
  cpu         = var.cli_cpu
  vcpu        = var.cli_cpu
  memory      = var.cli_memory_mb
  template_id = var.template_id
  group       = var.group_name

  disk {
    size   = var.cli_disk_mb
    target = "vda"
  }
}

#### The Ansible inventory file ############################################
resource "local_file" "ansible_inventory" {
  content  = templatefile("./templates/inventory.tmpl",
    {
      nodes-ips   = opennebula_virtual_machine.node[*].ip
      clients-ips = opennebula_virtual_machine.client[*].ip
      cli-ips     = opennebula_virtual_machine.cli[*].ip
    })
  filename = "../ansible/ansible_remote/inventory_dir/inventory"
}

### The Nodes Endpoints
resource "local_file" "nodes_endpoints" {
  content  = templatefile("./templates/endpoints.tmpl",
    {
      nodes-ips   = opennebula_virtual_machine.node[*].ip
      clients-ips = opennebula_virtual_machine.client[*].ip
      cli-ips     = opennebula_virtual_machine.cli[*].ip
    })
  filename = "../../configs/endpoints_remote.yml"
}

### The Nodes Endpoints for creating certificates
resource "local_file" "nodes_endpoints_certificate" {
  content  = templatefile("./templates/endpoints_certificate.tmpl",
    {
      nodes-ips   = opennebula_virtual_machine.node[*].ip
      clients-ips = opennebula_virtual_machine.client[*].ip
    })
  filename = "../../certificates/endpoints_remote"
}

# Output
output "public_ip_nodes" {
  value = opennebula_virtual_machine.node[*].ip
}

output "public_ip_clients" {
  value = opennebula_virtual_machine.client[*].ip
}

output "public_ip_cli" {
  value = opennebula_virtual_machine.cli[*].ip
}
