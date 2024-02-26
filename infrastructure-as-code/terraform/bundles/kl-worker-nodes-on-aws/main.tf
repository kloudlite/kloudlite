locals {
  check_k3s_join_token_is_set = {
    error_message = "k3s_join_token must be set"
    condition     = var.k3s_join_token != ""
  }

  check_k3s_server_public_dns_host_is_set = {
    error_message = "k3s_server_public_dns_host must be set"
    condition     = var.k3s_server_public_dns_host != ""
  }
}

resource "null_resource" "variable_validations" {
  lifecycle {
    precondition {
      error_message = local.check_k3s_join_token_is_set.error_message
      condition     = local.check_k3s_join_token_is_set.condition
    }

    precondition {
      error_message = local.check_k3s_server_public_dns_host_is_set.error_message
      condition     = local.check_k3s_server_public_dns_host_is_set.condition
    }
  }
}

module "constants" {
  source     = "../../modules/common/constants"
  depends_on = [null_resource.variable_validations]
}

module "ssh-rsa-key" {
  source     = "../../modules/common/ssh-rsa-key"
  depends_on = [null_resource.variable_validations]
}

resource "null_resource" "save_ssh_key" {
  count      = length(var.save_ssh_key_to_path) > 0 ? 1 : 0
  depends_on = [module.ssh-rsa-key]

  provisioner "local-exec" {
    quiet   = true
    command = "echo '${module.ssh-rsa-key.private_key}' > ${var.save_ssh_key_to_path} && chmod 600 ${var.save_ssh_key_to_path}"
  }
}

resource "random_id" "id" {
  byte_length = 4
  depends_on  = [null_resource.variable_validations]
}

resource "aws_key_pair" "k3s_worker_nodes_ssh_key" {
  key_name   = "${var.tracker_id}-${random_id.id.hex}-ssh-key"
  public_key = module.ssh-rsa-key.public_key
  depends_on = [null_resource.variable_validations]
}

locals {
  common_node_labels = {
    (module.constants.node_labels.kloudlite_release) : var.kloudlite_release,
    (module.constants.node_labels.provider_name) : "aws",
    (module.constants.node_labels.provider_region) : var.aws_region,
  }
}

module "aws-security-groups" {
  source     = "../../modules/aws/security-groups"
  depends_on = [null_resource.variable_validations]

  tracker_id = var.tracker_id
  vpc_id     = var.vpc.vpc_id

  create_for_k3s_workers = true

  allow_incoming_http_traffic = false
  expose_k8s_node_ports       = false
}

module "k3s-templates" {
  depends_on = [null_resource.variable_validations]
  source     = "../../modules/kloudlite/k3s/k3s-templates"
}

module "aws-amis" {
  source = "../../modules/aws/AMIs"
}

module "availability_zones" {
  source = "../../modules/aws/availability-zones"
}

module "aws-ec2-nodepool" {
  source     = "../../modules/aws/aws-ec2-nodepool"
  depends_on = [null_resource.variable_validations]
  for_each   = {for np_name, np_config in var.ec2_nodepools : np_name => np_config}

  tracker_id           = "${var.tracker_id}-${each.key}"
  ami                  = module.aws-amis.ubuntu_amd64_cpu_ami_id
  availability_zone    = each.value.availability_zone != "" ? each.value.availability_zone : module.availability_zones.names[0]
  iam_instance_profile = each.value.iam_instance_profile
  instance_type        = each.value.instance_type
  nvidia_gpu_enabled   = each.value.nvidia_gpu_enabled
  root_volume_size     = each.value.root_volume_size
  root_volume_type     = each.value.root_volume_type
  security_groups      = module.aws-security-groups.sg_for_k3s_agents_names
  ssh_key_name         = aws_key_pair.k3s_worker_nodes_ssh_key.key_name
  tags                 = var.tags
  vpc                  = {
    subnet_id              = var.vpc.vpc_public_subnet_ids[each.value.availability_zone != "" ? each.value.availability_zone : module.availability_zones.names[0]]
    vpc_security_group_ids = module.aws-security-groups.sg_for_k3s_agents_ids
  }
  nodes = {
    for name, cfg in each.value.nodes : name => {
      user_data_base64 = base64encode(templatefile(module.k3s-templates.k3s-agent-template-path, {
        kloudlite_config_directory = module.k3s-templates.kloudlite_config_directory

        vm_setup_script = templatefile(module.k3s-templates.k3s-vm-setup-template-path, {
          kloudlite_release          = var.kloudlite_release
          kloudlite_config_directory = module.k3s-templates.kloudlite_config_directory
        })

        tf_k3s_masters_dns_host = var.k3s_server_public_dns_host
        tf_k3s_token            = var.k3s_join_token
        tf_node_taints          = concat([],
          each.value.node_taints != null ? each.value.node_taints : [],
          each.value.nvidia_gpu_enabled == true ? module.constants.gpu_node_taints : [],
        )
        tf_node_labels = jsonencode(merge(
          local.common_node_labels,
          {
            (module.constants.node_labels.provider_az)   = each.value.availability_zone != "" ? each.value.availability_zone : module.availability_zones.names[0],
            (module.constants.node_labels.node_has_role) = "agent"
            (module.constants.node_labels.nodepool_name) : each.key,
            (module.constants.node_labels.provider_aws_instance_profile_name) : each.value.iam_instance_profile,
          },
          each.value.nvidia_gpu_enabled == true ? { (module.constants.node_labels.node_has_gpu) : "true" } : {}
        ))
        tf_node_name                 = "${each.key}-${name}"
        tf_use_cloudflare_nameserver = true
        tf_extra_agent_args          = var.extra_agent_args
      }))
      last_recreated_at = cfg.last_recreated_at
    }
  }
}

module "aws-spot-nodepool" {
  source                       = "../../modules/aws/aws-spot-nodepool"
  depends_on                   = [null_resource.variable_validations]
  for_each                     = {for np_name, np_config in var.spot_nodepools : np_name => np_config}
  tracker_id                   = "${var.tracker_id}-${each.key}"
  ami                          = module.aws-amis.ubuntu_amd64_cpu_ami_id
  availability_zone            = each.value.availability_zone != "" ? each.value.availability_zone : module.availability_zones.names[0]
  root_volume_size             = each.value.root_volume_size
  root_volume_type             = each.value.root_volume_type
  security_groups              = module.aws-security-groups.sg_for_k3s_agents_ids
  iam_instance_profile         = each.value.iam_instance_profile
  spot_fleet_tagging_role_name = each.value.spot_fleet_tagging_role_name
  ssh_key_name                 = aws_key_pair.k3s_worker_nodes_ssh_key.key_name
  cpu_node                     = each.value.cpu_node
  gpu_node                     = each.value.gpu_node
  vpc                          = {
    subnet_id              = var.vpc.vpc_public_subnet_ids[each.value.availability_zone != "" ? each.value.availability_zone : module.availability_zones.names[0]]
    vpc_security_group_ids = module.aws-security-groups.sg_for_k3s_agents_ids
  }
  nodes = {
    for name, cfg in each.value.nodes : name => {
      user_data_base64 = base64encode(templatefile(module.k3s-templates.k3s-agent-template-path, {
        kloudlite_config_directory = module.k3s-templates.kloudlite_config_directory

        vm_setup_script = templatefile(module.k3s-templates.k3s-vm-setup-template-path, {
          kloudlite_release          = var.kloudlite_release
          kloudlite_config_directory = module.k3s-templates.kloudlite_config_directory
        })

        tf_k3s_masters_dns_host = var.k3s_server_public_dns_host
        tf_k3s_token            = var.k3s_join_token
        tf_node_taints          = concat([],
          each.value.node_taints != null ? each.value.node_taints : [],
          each.value.gpu_node != null ? module.constants.gpu_node_taints : [],
        )
        tf_node_labels = jsonencode(merge(
          local.common_node_labels,
          {
            (module.constants.node_labels.provider_az)   = each.value.availability_zone != "" ? each.value.availability_zone : module.availability_zones.names[0],
            (module.constants.node_labels.node_has_role) = "agent"
            (module.constants.node_labels.node_is_spot)  = "true"
            (module.constants.node_labels.nodepool_name) : each.key,
            (module.constants.node_labels.provider_aws_instance_profile_name) : each.value.iam_instance_profile,
          },
          each.value.gpu_node != null ? { (module.constants.node_labels.node_has_gpu) : "true" } : {}
        ))
        tf_node_name                 = "${each.key}-${name}"
        tf_use_cloudflare_nameserver = true
        tf_extra_agent_args          = var.extra_agent_args
      }))
      last_recreated_at = cfg.last_recreated_at
    }
  }
}
