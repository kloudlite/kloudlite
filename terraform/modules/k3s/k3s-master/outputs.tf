output "k3s_server_token" {
  value = random_password.k3s_server_token.result
}

output "k3s_agent_token" {
  value = random_password.k3s_agent_token.result
}

output "k3s_primary_public_ip" {
  value = var.master_nodes[local.primary_master_node_name].public_ip
}

output "kubeconfig_with_public_ip" {
  value = base64encode(replace(base64decode(chomp(ssh_resource.create_revocable_kubeconfig[local.primary_master_node_name].result)), "https://127.0.0.1", "https://${var.master_nodes[local.primary_master_node_name].public_ip}"))
}

output "kubeconfig_with_public_host" {
  value = base64encode(replace(base64decode(chomp(ssh_resource.create_revocable_kubeconfig[local.primary_master_node_name].result)), "https://127.0.0.1", "https://${var.public_dns_host}"))
}
