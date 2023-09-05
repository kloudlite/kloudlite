output "k8s_masters_public_ips" {
  value = concat([aws_instance.k3s_primary_master.public_ip], [for instance in aws_instance.k3s_secondary_masters : instance.public_ip])
}

output "kubeconfig" {
  value = chomp(ssh_resource.grab_k8s_config.result)
}
