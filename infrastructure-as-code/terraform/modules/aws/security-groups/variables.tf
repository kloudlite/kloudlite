variable "allow_metrics_server_on_master" {
  description = "k3s metrics server: source: https://docs.k3s.io/installation/requirements#networking"
  type        = bool
}

variable "allow_incoming_http_traffic_on_master" {
  description = "enable http and https traffic into cluster"
  type        = bool
}

variable "expose_k8s_node_ports_on_master" {
  description = "expose k8s node ports to the internet"
  type        = bool
}

variable "allow_metrics_server_on_agent" {
  description = "k3s metrics server: source: https://docs.k3s.io/installation/requirements#networking"
  type        = bool
}

variable "allow_incoming_http_traffic_on_agent" {
  description = "enable http and https traffic into cluster"
  type        = bool
}

variable "expose_k8s_node_ports_on_agent" {
  description = "expose k8s node ports to the internet"
  type        = bool
}

variable "allow_outgoing_to_all_internet_on_agent" {
  description = "allow outgoing traffic to all internet on agent nodes"
  type        = bool
}
