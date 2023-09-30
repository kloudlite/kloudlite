output "k3s_masters" {
  value = {
    for node_name, v in merge(local.primary_master_nodes, local.secondary_master_nodes) :
    node_name => {
      public_ip         = module.ec2-nodes.ec2_instances_public_ip[node_name]
      availability_zone = local.nodes_config[node_name].az
    }
  }
}

output "k3s_token" {
  value = module.k3s-primary-master.k3s_token
}

output "kubeconfig" {
  value = module.k3s-primary-master.kubeconfig_with_public_host
}

output "kubeconfig_with_master_public_ip" {
  value = module.k3s-primary-master.kubeconfig_with_public_ip
}
