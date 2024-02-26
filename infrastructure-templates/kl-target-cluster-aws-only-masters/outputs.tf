output "k3s_masters" {
  value = module.kl-master-nodes-on-aws.k3s_masters
}

output "k3s_public_dns_host" {
  value = module.kl-master-nodes-on-aws.k3s_public_dns_host
}

output "k3s_server_token" {
  sensitive = true
  value     = module.kl-master-nodes-on-aws.k3s_server_token
}

output "k3s_agent_token" {
  sensitive = true
  value     = module.kl-master-nodes-on-aws.k3s_agent_token
}

output "kubeconfig" {
  sensitive = true
  value     = module.kl-master-nodes-on-aws.kubeconfig
}

output "kloudlite-k3s-params" {
  value = module.kl-master-nodes-on-aws.k3s-params
}

output "vpc_id" {
  value = module.kl-master-nodes-on-aws.vpc_id
}

output "vpc_public_subnets" {
  value = module.kl-master-nodes-on-aws.vpc_public_subnets
}
