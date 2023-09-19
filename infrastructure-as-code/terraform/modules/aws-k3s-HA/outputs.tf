output "ec2_instances" {
  value = module.ec2-nodes.ec2_instances
}

output "kubeconfig" {
  value = module.k3s-primary-master.kubeconfig_with_public_host
}

output "kubeconfig_with_master_public_ip" {
  value = module.k3s-primary-master.kubeconfig_with_public_ip
}


