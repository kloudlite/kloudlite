output "k3s_masters" {
  value = module.kl-master-nodes-on-aws.k3s_masters
}

output "k3s_agents" {
  value = module.kl-worker-nodes-on-aws.ec2-nodepools
}

output "kubeconfig" {
  sensitive = true
  value     = module.kl-master-nodes-on-aws.kubeconfig
}

output "vpc_id" {
  value = module.kl-master-nodes-on-aws.vpc_id
}

output "vpc_public_subnets" {
  value = module.kl-master-nodes-on-aws.vpc_public_subnets
}
