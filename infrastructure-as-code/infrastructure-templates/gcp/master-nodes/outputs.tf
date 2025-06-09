output "k3s_masters" {
  value = module.master-nodes-on-gcp.k3s_masters
}

output "k3s_public_dns_host" {
  value = module.master-nodes-on-gcp.k3s_public_dns_host
}

output "k3s_server_token" {
  sensitive = true
  value     = module.master-nodes-on-gcp.k3s_server_token
}

output "k3s_agent_token" {
  sensitive = true
  value     = module.master-nodes-on-gcp.k3s_agent_token
}

output "kubeconfig" {
  sensitive = true
  value     = module.master-nodes-on-gcp.kubeconfig
}

output "kloudlite-k3s-params" {
  value = module.master-nodes-on-gcp.k3s-params
}