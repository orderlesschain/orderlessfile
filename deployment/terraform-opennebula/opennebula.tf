# Work in Progress to migrate from OpenStack to OpenNebula
variable "nodes" {
  type    = list(string)
  default = [
    "172.24.33.3",
    "172.24.33.36",
  ]
}

variable "clients" {
  type    = list(string)
  default = [
    "172.24.33.134",
    "172.24.33.135",
  ]
}

variable "clis" {
  type    = list(string)
  default = ["172.24.33.167"]
}


#### The Ansible inventory file ############################################
resource "local_file" "ansible_inventory" {
  content = templatefile("./templates/inventory.tmpl",
    {
      nodes-ips   = var.nodes
      clients-ips = var.clients
      cli-ips     = var.clis
    })
  filename = "../ansible/ansible_remote/inventory_dir/inventory"
}

### The Nodes Endpoints
resource "local_file" "nodes_endpoints" {
  content = templatefile("./templates/endpoints.tmpl",
    {
      nodes-ips   = var.nodes
      clients-ips = var.clients
      cli-ips     = var.clis
    })
  filename = "../../configs/endpoints_remote.yml"
}

### The Nodes Endpoints for creating certificates
resource "local_file" "nodes_endpoints_certificate" {
  content = templatefile("./templates/endpoints_certificate.tmpl",
    {
      nodes-ips   = var.nodes
      clients-ips = var.clients
    })
  filename = "../../certificates/endpoints_remote"
}


