output "sg_for_k3s_masters_ids" {
  value = tolist([aws_security_group.k3s_master_sg.id])
}

output "sg_for_k3s_masters_names" {
  value = tolist([aws_security_group.k3s_master_sg.name])
}

output "sg_for_k3s_agents_ids" {
  value = tolist([aws_security_group.k3s_agent_sg.id])
}

output "sg_for_k3s_agents_names" {
  value = tolist([aws_security_group.k3s_agent_sg.name])
}
