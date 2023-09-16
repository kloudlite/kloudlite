output "security_group_k3s_masters_ids" {
  value = tolist([
    aws_security_group.k3s_server_nodes_requirements.id,
    aws_security_group.allow_metrics_server.id,
    aws_security_group.allows_ssh.id,
    aws_security_group.allows_incoming_http_traffic.id,
    aws_security_group.nodes_can_access_internet.id,
  ])
}

output "security_group_k3s_masters_names" {
  value = tolist([
    aws_security_group.k3s_server_nodes_requirements.name,
    aws_security_group.allow_metrics_server.name,
    aws_security_group.allows_ssh.name,
    aws_security_group.allows_incoming_http_traffic.name,
    aws_security_group.nodes_can_access_internet.name,
  ])
}

output "security_group_k3s_agents_ids" {
  value = tolist([
    aws_security_group.allows_ssh.id,
    aws_security_group.allow_metrics_server.id,
    aws_security_group.nodes_can_access_internet.id,
  ])
}

output "security_group_k3s_agents_names" {
  value = tolist([
    aws_security_group.allows_ssh.name,
    aws_security_group.allow_metrics_server.name,
    aws_security_group.nodes_can_access_internet.name,
  ])
}
