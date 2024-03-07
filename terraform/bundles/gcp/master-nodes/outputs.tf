output "public_ips" {
  value = {for name, cfg in module.master-nodes : name => cfg.public_ip}
}

output "k3s_masters" {
  value = {
    for name, cfg in module.master-nodes : name => {
      public_ip         = cfg.public_ip
      availability_zone = var.nodes[name].availability_zone
    }
  }
}

output "k3s_public_dns_host" {
  value = module.kloudlite-k3s-masters.k3s_public_dns_host
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

output "k3s-params" {
  value = module.kloudlite-k3s-masters.k3s-params
}
