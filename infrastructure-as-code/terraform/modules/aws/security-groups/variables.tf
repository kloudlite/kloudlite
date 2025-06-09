variable "tracker_id" {
  description = "tracker id"
  type        = string
}

variable "vpc_id" {
  type        = string
  description = "VPC Id onto which security groups will be created"
  default     = ""
}

variable "create_for_k3s_masters" {
  description = "create a group for k3s masters"
  type        = bool
  default     = false
}

variable "create_for_k3s_workers" {
  description = "create a group for k3s workers"
  type        = bool
  default     = false
}

variable "allow_incoming_http_traffic" {
  description = "enable http (80) and https (443) traffic into cluster"
  type        = bool
}

variable "expose_k8s_node_ports" {
  description = "expose k8s node ports to the internet, to allow incoming traffic on any port"
  type        = bool
}