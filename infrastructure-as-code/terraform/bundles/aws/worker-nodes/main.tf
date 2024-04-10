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

    precondition {
      error_message = "either ec2_nodepool or spot_nodepool must be set"
      condition     = anytrue([
        var.ec2_nodepool == null || var.spot_nodepool != null,
        var.ec2_nodepool != null || var.spot_nodepool == null,
      ])
    }

    precondition {
      error_message = "a spot nodepool can only consist of one kind of node either cpu only or gpu only"
      condition     = anytrue([
        var.spot_nodepool == null,
        var.spot_nodepool == null ? false : var.spot_nodepool.cpu_node == null && var.spot_nodepool.gpu_node != null,
        var.spot_nodepool == null ? false : var.spot_nodepool.cpu_node != null && var.spot_nodepool.gpu_node == null,
      ])
    }
  }
}

module "constants" {
  source     = "../../../modules/common/constants"
  depends_on = [null_resource.variable_validations]
}

module "ssh-rsa-key" {
  source     = "../../../modules/common/ssh-rsa-key"
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
  source     = "../../../modules/aws/security-groups"
  depends_on = [null_resource.variable_validations]

  tracker_id = var.tracker_id
  vpc_id     = var.vpc_id

  create_for_k3s_workers = true

  allow_incoming_http_traffic = false
  expose_k8s_node_ports       = false
}

module "k3s-templates" {
  depends_on = [null_resource.variable_validations]
  source     = "../../../modules/kloudlite/k3s/k3s-templates"
}

module "aws-amis" {
  source = "../../../modules/aws/AMIs"
}

module "aws-ec2-nodepool" {
  count                = var.ec2_nodepool != null ? 1 : 0
  source               = "../../../modules/aws/aws-ec2-nodepool"
  depends_on           = [null_resource.variable_validations, module.aws-security-groups.sg_for_k3s_agents_ids]
  tracker_id           = "${var.tracker_id}-${var.nodepool_name}"
  ami                  = module.aws-amis.ubuntu_amd64_cpu_ami_id
  availability_zone    = var.availability_zone
  iam_instance_profile = var.iam_instance_profile
  instance_type        = var.ec2_nodepool.instance_type
  nvidia_gpu_enabled   = var.nvidia_gpu_enabled
  root_volume_size     = var.ec2_nodepool.root_volume_size
  root_volume_type     = var.ec2_nodepool.root_volume_type
  security_groups      = module.aws-security-groups.sg_for_k3s_agents_names
  ssh_key_name         = aws_key_pair.k3s_worker_nodes_ssh_key.key_name
  tags                 = var.tags
  vpc                  = {
    subnet_id              = var.vpc_subnet_id
    vpc_security_group_ids = module.aws-security-groups.sg_for_k3s_agents_ids
  }
  nodes = {
    for name, cfg in var.ec2_nodepool.nodes : name => {
      user_data_base64 = base64encode(templatefile(module.k3s-templates.k3s-agent-template-path, {
        kloudlite_config_directory = module.k3s-templates.kloudlite_config_directory

        vm_setup_script = templatefile(module.k3s-templates.k3s-vm-setup-template-path, {
          kloudlite_release          = var.kloudlite_release
          kloudlite_config_directory = module.k3s-templates.kloudlite_config_directory
        })

        tf_k3s_masters_dns_host = var.k3s_server_public_dns_host
        tf_k3s_token            = var.k3s_join_token
        tf_node_taints          = concat([],
          var.node_taints != null ? var.node_taints : [],
          var.nvidia_gpu_enabled == true ? module.constants.gpu_node_taints : [],
        )
        tf_node_labels = jsonencode(merge(
          local.common_node_labels,
          {
            (module.constants.node_labels.provider_az)   = var.availability_zone
            (module.constants.node_labels.node_has_role) = "agent"
            (module.constants.node_labels.nodepool_name) : var.nodepool_name
            (module.constants.node_labels.provider_aws_instance_profile_name) : var.iam_instance_profile
          },
          var.nvidia_gpu_enabled == true ? { (module.constants.node_labels.node_has_gpu) : "true" } : {}
        ))
        tf_node_name                 = "${var.nodepool_name}-${name}"
        tf_use_cloudflare_nameserver = true
        tf_extra_agent_args          = var.extra_agent_args
      }))
      last_recreated_at = cfg.last_recreated_at
    }
  }
}

module "aws-spot-nodepool" {
  source     = "../../../modules/aws/aws-spot-nodepool"
  depends_on = [
    null_resource.variable_validations, module.aws-security-groups.sg_for_k3s_agents_ids
  ]
  count                        = var.spot_nodepool != null ? 1 : 0
  tracker_id                   = "${var.tracker_id}-${var.nodepool_name}"
  ami                          = module.aws-amis.ubuntu_amd64_cpu_ami_id
  availability_zone            = var.availability_zone
  root_volume_size             = var.spot_nodepool.root_volume_size
  root_volume_type             = var.spot_nodepool.root_volume_type
  security_groups              = module.aws-security-groups.sg_for_k3s_agents_ids
  iam_instance_profile         = var.iam_instance_profile
  spot_fleet_tagging_role_name = var.spot_nodepool.spot_fleet_tagging_role_name
  ssh_key_name                 = aws_key_pair.k3s_worker_nodes_ssh_key.key_name
  cpu_node                     = var.spot_nodepool.cpu_node
  gpu_node                     = var.spot_nodepool.gpu_node
  vpc                          = {
    subnet_id              = var.vpc_subnet_id
    vpc_security_group_ids = module.aws-security-groups.sg_for_k3s_agents_ids
  }
  nodes = {
    for name, cfg in var.spot_nodepool.nodes : name => {
      user_data_base64 = base64encode(templatefile(module.k3s-templates.k3s-agent-template-path, {
        kloudlite_config_directory = module.k3s-templates.kloudlite_config_directory

        vm_setup_script = templatefile(module.k3s-templates.k3s-vm-setup-template-path, {
          kloudlite_release          = var.kloudlite_release
          kloudlite_config_directory = module.k3s-templates.kloudlite_config_directory
        })

        tf_k3s_masters_dns_host = var.k3s_server_public_dns_host
        tf_k3s_token            = var.k3s_join_token
        tf_node_taints          = concat([],
          var.node_taints != null ? var.node_taints : [],
          var.spot_nodepool.gpu_node != null ? module.constants.gpu_node_taints : [],
        )
        tf_node_labels = jsonencode(merge(
          local.common_node_labels,
          {
            (module.constants.node_labels.provider_az)                        = var.availability_zone
            (module.constants.node_labels.node_has_role)                      = "agent"
            (module.constants.node_labels.node_is_spot)                       = "true"
            (module.constants.node_labels.nodepool_name)                      = var.nodepool_name
            (module.constants.node_labels.provider_aws_instance_profile_name) = var.iam_instance_profile
          },
          var.spot_nodepool.gpu_node != null ? { (module.constants.node_labels.node_has_gpu) : "true" } : {}
        ))
        tf_node_name                 = "${var.nodepool_name}-${name}"
        tf_use_cloudflare_nameserver = true
        tf_extra_agent_args          = var.extra_agent_args
      }))
      last_recreated_at = cfg.last_recreated_at
    }
  }
}
