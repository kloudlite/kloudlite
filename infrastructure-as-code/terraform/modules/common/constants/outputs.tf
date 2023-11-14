locals {
  kloudlite_node_labels = {
    provider_name          = "kloudlite.io/provider.name"
    provider_region        = "kloudlite.io/provider.region"
    provider_az            = "kloudlite.io/provider.az"
    provider_instance_type = "kloudlite.io/provider.instance-type"

    provider_aws_instance_profile_name = "kloudlite.io/provider.aws.instance-profile-name"

    nodepool_name = "kloudlite.io/nodepool.name"
    node_has_role = "kloudlite.io/node.has-role"
    node_has_gpu  = "kloudlite.io/node.has-gpu"
    node_is_spot  = "kloudlite.io/node.is-spot"
  }

  k8s_default_node_labels = {
    master_node = "node-role.kubernetes.io/master"
  }
}

output "master_node_taints" {
  value = [
    {
      key    = local.k8s_default_node_labels.master_node
      effect = "NoSchedule",
      value  = ""
    }
  ]
}

output "master_node_tolerations" {
  value = [
    {
      key      = local.k8s_default_node_labels.master_node
      operator = "Exists"
      effect   = "NoSchedule"
    }
  ]
}

output "gpu_node_taints" {
  value = [
    {
      key    = local.kloudlite_node_labels.node_has_gpu
      effect = "NoSchedule"
      value  = "true"
    }
  ]
}

output "gpu_node_tolerations" {
  value = [
    {
      key      = local.kloudlite_node_labels.node_has_gpu
      operator = "Equal"
      effect   = "NoSchedule"
      value    = "true"
    }
  ]
}

output "master_node_selector" {
  value = {
    (local.k8s_default_node_labels.master_node) = "true"
  }
}

output "gpu_node_selector" {
  value = {
    (local.kloudlite_node_labels.node_has_gpu) = "true"
  }
}

output "agent_node_selector" {
  value = {
    (local.kloudlite_node_labels.node_has_role) = "agent"
  }
}

output "spot_node_selector" {
  value = {
    (local.kloudlite_node_labels.node_is_spot) = "true"
  }
}

output "node_labels" {
  value = local.kloudlite_node_labels
}