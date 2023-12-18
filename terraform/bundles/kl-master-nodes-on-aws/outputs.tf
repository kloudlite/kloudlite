output "k3s_masters" {
  value = {
    for node_name, output in module.k3s-master-instances : node_name => {
      public_ip         = output.public_ip
      availability_zone = var.k3s_masters.nodes[node_name].availability_zone
    }
  }
}

output "k3s_public_dns_host" {
  value = var.k3s_masters.public_dns_host
}

output "k3s_server_token" {
  sensitive = true
  value     = module.kloudlite-k3s-masters.k3s_server_token
}

output "k3s_agent_token" {
  sensitive = true
  value     = module.kloudlite-k3s-masters.k3s_agent_token
}

output "kubeconfig" {
  sensitive = true
  value     = module.kloudlite-k3s-masters.kubeconfig
}