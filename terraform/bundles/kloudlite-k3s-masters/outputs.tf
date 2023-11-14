output "k3s_masters" {
  value = {
    for node_name, cfg in var.master_nodes : node_name => {
      public_ip         = cfg.public_ip
      availability_zone = cfg.availability_zone
    }
  }
}

output "k3s_primary_master_public_ip" {
  value = module.k3s-masters.k3s_primary_public_ip
}

output "k3s_public_dns_host" {
  value = var.public_dns_host
}

output "k3s_server_token" {
  sensitive = true
  value     = module.k3s-masters.k3s_server_token
}

output "k3s_agent_token" {
  sensitive = true
  value     = module.k3s-masters.k3s_agent_token
}

output "kubeconfig" {
  sensitive = true
  value     = module.k3s-masters.kubeconfig_with_public_host
}

output "kloudlite_namespace" {
  value = local.kloudlite_namespace
}