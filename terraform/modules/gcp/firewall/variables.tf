variable "network_name" {
  type        = string
  description = "GCP Network name"
}

variable "name_prefix" {
  type = string
}

variable "for_master_nodes" {
  type        = bool
  description = "firewall for master nodes ?"
  default     = false
}

variable "for_worker_nodes" {
  type        = bool
  description = "firewall for worker nodes ?"
  default     = false
}

variable "for_vm_group" {
  type        = bool
  description = "firewall for vm group ?"
  default     = false
}

variable "target_tags" {
  type        = list(string)
  description = "tags of VMs over which this firewall rule should apply"
}

// firewall rules

variable "allow_incoming_http_traffic" {
  type        = bool
  description = "allow incoming http traffic"
}

variable "allow_node_ports" {
  type        = bool
  description = "should allow node ports ?"
}

variable "allow_ssh" {
  type        = bool
  description = "should allow node ports ?"
}
