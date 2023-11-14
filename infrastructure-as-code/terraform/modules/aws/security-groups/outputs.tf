output "sg_for_k3s_masters_ids" {
  value = var.create_group_for_k3s_masters ?  tolist([aws_security_group.k3s_master_sg[0].id]) : []
}

output "sg_for_k3s_masters_names" {
  value = var.create_group_for_k3s_masters ?  tolist([aws_security_group.k3s_master_sg[0].name]) : []
}

output "sg_for_k3s_agents_ids" {
  value = var.create_group_for_k3s_workers ? tolist([aws_security_group.k3s_agent_sg[0].id]) : []
}

output "sg_for_k3s_agents_names" {
  value = var.create_group_for_k3s_workers ? tolist([aws_security_group.k3s_agent_sg[0].name]) : []
}
