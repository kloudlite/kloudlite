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